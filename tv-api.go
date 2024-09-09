package main

import (
	"fmt"
	"maps"
	"net/http"
	"slices"
	"time"

	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

var ws *websocket.Conn // websocket connection

func OpenConnection() error {
	url := "wss://data.tradingview.com/socket.io/websocket"
	header := http.Header{}
	header.Add("Origin", "https://www.tradingview.com")

	var err error
	ws, _, err = websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		return err
	}

	// separate thread for the msgs
	go func() { // TODO make things thread-safe
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				fmt.Println("Read error: ", err)
				return // quit reading if error
			}

			readMessage(string(message))
		}
	}()

	return auth()
}

func AddRealtimeSymbols(symbols []string) error {
	symbols_conv := convertStringArrToInterfaceArr(symbols)
	if err := sendMessage("quote_add_symbols", append([]interface{}{qssq}, symbols_conv...)); err != nil {
		return err
	}

	for _, symbol := range symbols {
		realtimeSymbols[symbol] = true
	}

	return quoteFastSymbols()
}

func RemoveRealtimeSymbols(symbols []string) error {
	symbols_conv := convertStringArrToInterfaceArr(symbols)
	if err := sendMessage("quote_remove_symbols", append([]interface{}{qssq}, symbols_conv...)); err != nil {
		return err
	}

	for _, symbol := range symbols {
		delete(realtimeSymbols, symbol)
	}

	return quoteFastSymbols()
}

func RequestMoreData(candleCount int) error {
	time.Sleep(100 * time.Millisecond) //removing this stops history and requestMoreData from printing for some reason. Figure out why

	/*
		if err := sendMessage("request_more_data", append([]interface{}{csToken}, "s1", candleCount)); err != nil {
			return err
		}
		return nil*/

	return sendMessage("request_more_data", append([]interface{}{csToken}, "s1", candleCount))
}

func GetHistory(symbol string, timeframe string, sessionType string) error {
	err := resolveSymbol(symbol, sessionType)
	if err != nil {
		return err
	}

	seriesCounter++
	series := "s" + strconv.Itoa(seriesCounter)
	id := resolvedSymbols[symbol]

	seriesMap[series] = symbol
	if !seriesCreated {
		seriesCreated = true

		err := sendMessage("create_series", []interface{}{csToken, "s1", series, id, timeframe, initHistoryCandles, ""})
		if err != nil {
			return err
		}
	} else {
		err := sendMessage("modify_series", []interface{}{csToken, "s1", series, id, timeframe, ""})
		if err != nil {
			return err
		}
	}

	return nil
}

func SwitchTimezone(timezone string) error {
	return sendMessage("switch_timezone", append([]interface{}{csToken}, timezone))
}

func quoteFastSymbols() error {
	symbols := slices.Collect(maps.Keys(realtimeSymbols))
	symbols_conv := convertStringArrToInterfaceArr(symbols)

	return sendMessage("quote_fast_symbols", append([]interface{}{qs}, symbols_conv...))
}

func auth() error {
	authMsgs := []struct {
		name string
		args []interface{}
	}{
		{"set_auth_token", []interface{}{"unauthorized_user_token"}},
		{"chart_create_session", []interface{}{csToken, ""}},
		{"quote_create_session", []interface{}{qs}},
		{"quote_create_session", []interface{}{qssq}},
		{"quote_set_fields", []interface{}{qssq, "base-currency-logoid", "ch", "chp", "currency-logoid", "currency_code", "currency_id", "base_currency_id", "current_session", "description", "exchange", "format", "fractional", "is_tradable", "language", "local_description", "listed_exchange", "logoid", "lp", "lp_time", "minmov", "minmove2", "original_name", "pricescale", "pro_name", "short_name", "type", "typespecs", "update_mode", "volume", "variable_tick_size", "value_unit_id"}},
	}

	for _, token := range authMsgs {
		if err := sendMessage(token.name, token.args); err != nil {
			return err
		}
	}

	return nil
}

func sendMessage(name string, args []interface{}) error {
	if ws == nil {
		return errors.New("websocket is null")
	}
	/* debugging
	fmt.Println("name : ", name)
	for _, arg := range args {
		fmt.Print(arg, ",")
	}
	fmt.Println()*/
	message, err := json.Marshal(
		map[string]interface{}{
			"m": name,
			"p": args,
		},
	)

	if err != nil {
		return err
	}

	err = ws.WriteMessage(websocket.TextMessage, []byte("~m~"+strconv.Itoa(len(message))+"~m~"+string(message)))
	if err != nil {
		return err
	}

	return nil
}

func readMessage(buffer string) { // TODO better error handling
	msgs := strings.Split(buffer, "~m~")
	for _, msg := range msgs {
		var res map[string]interface{}
		err := json.Unmarshal([]byte(msg), &res)

		if err != nil {
			// not json
			if strings.Contains(msg, "~h~") {
				err = ws.WriteMessage(websocket.TextMessage, []byte("~m~"+strconv.Itoa(len(msg))+"~m~"+msg))

				if err != nil {
					fmt.Println(err) // print error, TODO but continue anyway? or crash?
				}
			}

			continue
		}

		// is json
		if res["m"] == "qsd" { // realtime price changes
			resp, ok := res["p"].([]interface{})
			if !ok {
				continue
			}

			info, ok := resp[1].(map[string]interface{})
			if !ok {
				continue
			}

			fmt.Println("symbol: ", info["n"]) // TODO actually do something
			if data, ok := info["v"].(map[string]interface{}); ok {
				fmt.Println("volume: ", data["volume"])
				fmt.Println("current price: ", data["lp"])
				fmt.Println("change in price: ", data["ch"])
				fmt.Println("change in price %:", data["chp"])
				fmt.Println("the timestamp: ", data["lp_time"])
			}
		} else if res["m"] == "timescale_update" { // get historical data
			resp, ok := res["p"].([]interface{})
			if !ok {
				continue
			}

			info, ok := resp[1].(map[string]interface{})
			if !ok {
				continue
			}

			seriesInfo, ok := info["s1"].(map[string]interface{})
			if !ok {
				continue
			}

			allData, ok := seriesInfo["s"].([]interface{})
			if !ok {
				continue
			}

			seriesId, ok := seriesInfo["t"].(string)
			if !ok {
				continue
			}

			// TODO actually do something
			fmt.Println("symbol: ", seriesMap[seriesId])
			for _, dataElement := range allData {
				dataElement, ok := dataElement.(map[string]interface{})
				if !ok {
					continue
				}

				data, ok := dataElement["v"].([]interface{})
				if !ok {
					continue
				}
				fmt.Println("the timestamp: ", data[0])
				fmt.Println("open: ", data[1])
				fmt.Println("high: ", data[2])
				fmt.Println("low: ", data[3])
				fmt.Println("close: ", data[4])
				if len(data) >= 6 {
					fmt.Println("volume: ", data[5])
				}
			}
		}
	}
}