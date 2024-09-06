package main

import (
	"fmt"
	"net/http"

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
	go func() {
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				fmt.Println("Read error: ", err)
				return // quit reading if error
			}

			readMessage(string(message))
		}
	}()

	if err := auth(); err != nil {
		return err
	}

	return nil
}

func AddRealtimeSymbols(symbols []string) error {
	if err := sendMessage("quote_add_symbols", append([]string{qssq}, symbols...)); err != nil {
		return err
	}

	if err := sendMessage("quote_fast_symbols", append([]string{qs}, symbols...)); err != nil {
		return err
	}

	return nil
}

func RemoveRealtimeSymbols(symbols []string) error {
    if err := sendMessage("quote_remove_symbols", append([]string{qssq}, symbols...)); err != nil {
		return err
	}

	return nil
}

func auth() error {
	authMsgs := []struct {
		name string
		args    []string
	}{
		{"set_auth_token", []string{"unauthorized_user_token"}},
		{"chart_create_session", []string{csToken, ""}},
		{"quote_create_session", []string{qs}},
		{"quote_create_session", []string{qssq}},
		{"quote_set_fields", []string{qssq, "base-currency-logoid", "ch", "chp", "currency-logoid", "currency_code", "currency_id", "base_currency_id", "current_session", "description", "exchange", "format", "fractional", "is_tradable", "language", "local_description", "listed_exchange", "logoid", "lp", "lp_time", "minmov", "minmove2", "original_name", "pricescale", "pro_name", "short_name", "type", "typespecs", "update_mode", "volume", "variable_tick_size", "value_unit_id"}},
	}

	for _, token := range authMsgs {
		if err := sendMessage(token.name, token.args); err != nil {
			return err
		}
	}

	return nil
}

func sendMessage(name string, args []string) error {
	if ws == nil {
		return errors.New("websocket is null")
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

	err = ws.WriteMessage(websocket.TextMessage, []byte("~m~" + strconv.Itoa(len(message)) + "~m~" + string(message)))
	if err != nil {
		return err
	}

	return nil
}

func readMessage(buffer string) {
	msgs := strings.Split(buffer, "~m~")
	for _, msg := range msgs {
		var res map[string]interface{}
		err := json.Unmarshal([]byte(msg), &res)

		if err != nil {
			// not json
			if strings.Contains(msg, "~h~") {
				err = ws.WriteMessage(websocket.TextMessage, []byte("~m~" + strconv.Itoa(len(msg)) + "~m~" + msg))

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

			fmt.Println("symbol: ", info["n"])
			if data, ok := info["v"].(map[string]interface{}); ok {
				fmt.Println("volume: ", data["volume"])
				fmt.Println("current price: ", data["lp"])
				fmt.Println("change in price: ", data["ch"])
				fmt.Println("change in price %:", data["chp"])
				fmt.Println("the timestamp: ", data["lp_time"])
			}
		} else if res["m"] == "timescale_update" { // get historical data
			// TODO
		}
	}
}
