package main

import (
	"fmt"
	"maps"
	"slices"
	"sync"
	"time"

	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

var mu Container

type Container struct { //mutex + check for state
	mutex    sync.Mutex
	isLocked bool
}

func Unlock() bool { //unlocking mutex
	if !mu.isLocked { //failure: mutex is already unlocked
		return false
	}
	mu.isLocked = false
	mu.mutex.Unlock()
	return true
}

func Lock() bool { //locking mutex
	if mu.isLocked { //failure: mutex is already locked
		return false
	}
	mu.isLocked = true
	mu.mutex.Lock()
	return true
}

func (tv_api *TV_API) AddRealtimeSymbols(symbols []string) error {
	symbols_conv := convertStringArrToInterfaceArr(symbols)
	if err := tv_api.sendMessage("quote_add_symbols", append([]interface{}{qssq}, symbols_conv...)); err != nil {
		return err
	}

	for _, symbol := range symbols {
		realtimeSymbols[symbol] = true
	}

	return tv_api.quoteFastSymbols()
}

func (tv_api *TV_API) RemoveRealtimeSymbols(symbols []string) error {
	symbols_conv := convertStringArrToInterfaceArr(symbols)
	if err := tv_api.sendMessage("quote_remove_symbols", append([]interface{}{qssq}, symbols_conv...)); err != nil {
		return err
	}

	for _, symbol := range symbols {
		delete(realtimeSymbols, symbol)
	}

	return tv_api.quoteFastSymbols()
}

func (tv_api *TV_API) RequestMoreData(candleCount int) error {
	err := tv_api.sendMessage("request_more_data", append([]interface{}{csToken}, HISTORY_TOKEN, candleCount))
	if err != nil {
		return err
	}
	return waitForMessage(5000 + candleCount)
	//TODO: make this more efficient; possibly apply log math? I'm thinking since requesting an absurd amount of candles (e.g, 100k+), if an issue occurs it may take absurd amounts of time to just note a timeout.
	//same for history; maybe pass some paa
}

func (tv_api *TV_API) GetHistory(symbol string, timeframe Timeframe, sessionType SessionType) error {
	err := tv_api.resolveSymbol(symbol, sessionType)
	if err != nil {
		return err
	}

	seriesCounter++
	series := "s" + strconv.FormatUint(seriesCounter, 10)
	id := resolvedSymbols[symbol]

	seriesMap[series] = symbol

	if !seriesCreated {
		seriesCreated = true
		err := tv_api.sendMessage("create_series", []interface{}{csToken, HISTORY_TOKEN, series, id, string(timeframe), initHistoryCandles, ""})
		if err != nil {
			return err
		}
		err = waitForMessage(5000)
		if err != nil {
			return err
		}
	} else {
		err := tv_api.sendMessage("modify_series", []interface{}{csToken, HISTORY_TOKEN, series, id, string(timeframe), ""})
		if err != nil {
			return err
		}
		err = waitForMessage(5000)
		if err != nil {
			return err
		}
	}

	return nil
}

func waitForMessage(maxWait int) error { //Please replace; is just a waiter for a message

	var lockstate = Lock()

	if !lockstate {
		return errors.New("waitForMesssage: mutex is already locked")
	}

	start := time.Now()
	for mu.isLocked && int(time.Since(start).Milliseconds()) <= maxWait {

	}

	lockstate = Unlock()
	if lockstate {
		return errors.New("waitForMessage: timeout on waiting for message")
	}
	return nil
}

func (tv_api *TV_API) SwitchTimezone(timezone string) error {
	return tv_api.sendMessage("switch_timezone", append([]interface{}{csToken}, timezone))
}

func (tv_api *TV_API) quoteFastSymbols() error {
	symbols := slices.Collect(maps.Keys(realtimeSymbols))
	symbols_conv := convertStringArrToInterfaceArr(symbols)

	return tv_api.sendMessage("quote_fast_symbols", append([]interface{}{qs}, symbols_conv...))
}

func (tv_api *TV_API) auth() error {
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
		if err := tv_api.sendMessage(token.name, token.args); err != nil {
			return err
		}
	}

	return nil
}

func (tv_api *TV_API) sendMessage(name string, args []interface{}) error {
	if tv_api.ws == nil {
		return errors.New("sendMessage: websocket is null")
	}

	message, err := json.Marshal(
		map[string]interface{}{
			"m": name,
			"p": args,
		},
	)

	if err != nil {
		return err
	}

	err = tv_api.ws.WriteMessage(websocket.TextMessage, []byte(SEPARATOR+strconv.Itoa(len(message))+SEPARATOR+string(message)))
	if err != nil {
		return err
	}

	return nil
}

func (tv_api *TV_API) readMessage(buffer string) error { // TODO better error handling
	msgs := strings.Split(buffer, "~m~")
	for _, msg := range msgs {
		var res map[string]interface{}
		err := json.Unmarshal([]byte(msg), &res)

		if err != nil {
			// not json
			if strings.Contains(msg, "~h~") {
				err = tv_api.ws.WriteMessage(websocket.TextMessage, []byte(SEPARATOR+strconv.Itoa(len(msg))+SEPARATOR+msg))

				if err != nil {
					return err
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

			fmt.Println("symbol: ", info["n"])                      // TODO actually do something
			if data, ok := info["v"].(map[string]interface{}); ok { // TODO some of these vals can be null, add a check for that
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

			seriesInfo, ok := info[HISTORY_TOKEN].(map[string]interface{})
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
		} else if res["m"] == "series_completed" {
			Unlock()
		} else if res["m"] == "critical_error" {
			return errors.New("readMessage: TradingView Critical Error: " + msg)
		} else if res["m"] == "protocol_error" {
			return errors.New("readMessage: TradingView Protocol Error: " + msg)
		}
	}
	return nil
}
