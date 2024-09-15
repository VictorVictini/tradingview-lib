package tradingview

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

func (api *API) RemoveRealtimeSymbols(symbols []string) error {
	symbols_conv := convertInterfaceArr(symbols)
	if err := api.sendWriteThread("quote_remove_symbols", append([]interface{}{api.session.quote.symbolQuotes}, symbols_conv...)); err != nil {
		return err
	}

	for _, symbol := range symbols {
		delete(api.symbols.resolvedIDs, symbol)
	}

	return api.updateRealtimeSymbols()
}

func (api *API) SwitchTimezone(timezone string) error {
	return api.sendWriteThread("switch_timezone", append([]interface{}{api.session.chart.key}, timezone))
}

func (api *API) auth() error {
	authMsgs := []request{
		{"set_auth_token", []interface{}{"unauthorized_user_token"}},
		{"chart_create_session", []interface{}{api.session.chart.key, ""}},
		{"quote_create_session", []interface{}{api.session.quote.key}},
		{"quote_create_session", []interface{}{api.session.quote.symbolQuotes}},
		{"quote_set_fields", []interface{}{api.session.quote.symbolQuotes, "base-currency-logoid", "ch", "chp", "currency-logoid", "currency_code", "currency_id", "base_currency_id", "current_session", "description", "exchange", "format", "fractional", "is_tradable", "language", "local_description", "listed_exchange", "logoid", "lp", "lp_time", "minmov", "minmove2", "original_name", "pricescale", "pro_name", "short_name", "type", "typespecs", "update_mode", "volume", "variable_tick_size", "value_unit_id"}},
	}

	for _, token := range authMsgs {
		if err := api.sendWriteThread(token.name, token.args); err != nil {
			return err
		}
	}

	return nil
}

func (api *API) sendServer(name string, args []interface{}) error {
	if api.ws == nil {
		return errors.New("sendServer: websocket is null")
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

	err = api.ws.WriteMessage(websocket.TextMessage, []byte(SEPARATOR+strconv.Itoa(len(message))+SEPARATOR+string(message)))
	if err != nil {
		return err
	}

	return nil
}

func (api *API) readServerMessage(buffer string) error {
	msgs := strings.Split(buffer, "~m~")
	for _, msg := range msgs {
		var res map[string]interface{}
		err := json.Unmarshal([]byte(msg), &res)

		if err != nil {
			// not json
			if strings.Contains(msg, "~h~") {
				err = api.ws.WriteMessage(websocket.TextMessage, []byte(SEPARATOR+strconv.Itoa(len(msg))+SEPARATOR+msg))

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

			var res map[string]interface{} = make(map[string]interface{})
			res["symbol"] = info["n"]
			if data, ok := info["v"].(map[string]interface{}); ok {
				res["volume"] = data["volume"]
				res["current_price"] = data["lp"]
				res["price_change"] = data["ch"]
				res["price_change_percentage"] = data["chp"]
				res["timestamp"] = data["lp_time"]
			}

			api.Channels.Read <- res
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

			var res map[string]interface{} = make(map[string]interface{})
			res["symbol"] = api.series.mapsSymbols[seriesId]
			var timestamp, open, high, low, close, volume []interface{}
			for _, dataElement := range allData {
				dataElement, ok := dataElement.(map[string]interface{})
				if !ok {
					continue
				}

				data, ok := dataElement["v"].([]interface{})
				if !ok {
					continue
				}

				timestamp = append(timestamp, data[0])
				open = append(open, data[1])
				high = append(high, data[2])
				low = append(low, data[3])
				close = append(close, data[4])
				if len(data) >= 6 {
					volume = append(volume, data[5])
				} else {
					volume = append(volume, nil)
				}
			}

			res["timestamp"] = timestamp
			res["open"] = open
			res["high"] = high
			res["low"] = low
			res["close"] = close
			res["volume"] = volume

			api.Channels.Read <- res
		} else if api.halted.on != "" && res["m"] == api.halted.on {
			api.halted.mutex.Unlock()
			api.halted.on = ""
		} else if res["m"] == "critical_error" {
			return errors.New("readServerMessage: TradingView Critical Error: " + msg)
		} else if res["m"] == "protocol_error" {
			return errors.New("readServerMessage: TradingView Protocol Error: " + msg)
		}
	}
	return nil
}
