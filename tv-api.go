package tradingview

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

func (tv_api *TV_API) RemoveRealtimeSymbols(symbols []string) error {
	symbols_conv := convertStringArrToInterfaceArr(symbols)
	if err := tv_api.sendToWriteChannel("quote_remove_symbols", append([]interface{}{tv_api.qssq}, symbols_conv...)); err != nil {
		return err
	}

	for _, symbol := range symbols {
		delete(tv_api.realtimeSymbols, symbol)
	}

	return tv_api.quoteFastSymbols()
}

func (tv_api *TV_API) RequestMoreData(candleCount int) error {
	return tv_api.sendToWriteChannel("request_more_data", append([]interface{}{tv_api.csToken}, HISTORY_TOKEN, candleCount))
}

func (tv_api *TV_API) GetHistory(symbol string, timeframe Timeframe, sessionType SessionType) error {
	err := tv_api.resolveSymbol(symbol, sessionType)
	if err != nil {
		return err
	}

	tv_api.seriesCounter++
	series := "s" + strconv.FormatUint(tv_api.seriesCounter, 10)
	id := tv_api.resolvedSymbols[symbol]

	tv_api.seriesMap[series] = symbol

	// possibly use sync.Once?
	if !tv_api.seriesCreated {
		tv_api.seriesCreated = true
		return tv_api.sendToWriteChannel("create_series", []interface{}{tv_api.csToken, HISTORY_TOKEN, series, id, string(timeframe), INITIAL_HISTORY_CANDLES, ""})
	}
	return tv_api.sendToWriteChannel("modify_series", []interface{}{tv_api.csToken, HISTORY_TOKEN, series, id, string(timeframe), ""})
}

func (tv_api *TV_API) SwitchTimezone(timezone string) error {
	return tv_api.sendToWriteChannel("switch_timezone", append([]interface{}{tv_api.csToken}, timezone))
}

func (tv_api *TV_API) auth() error {
	authMsgs := []struct {
		name string
		args []interface{}
	}{
		{"set_auth_token", []interface{}{"unauthorized_user_token"}},
		{"chart_create_session", []interface{}{tv_api.csToken, ""}},
		{"quote_create_session", []interface{}{tv_api.qs}},
		{"quote_create_session", []interface{}{tv_api.qssq}},
		{"quote_set_fields", []interface{}{tv_api.qssq, "base-currency-logoid", "ch", "chp", "currency-logoid", "currency_code", "currency_id", "base_currency_id", "current_session", "description", "exchange", "format", "fractional", "is_tradable", "language", "local_description", "listed_exchange", "logoid", "lp", "lp_time", "minmov", "minmove2", "original_name", "pricescale", "pro_name", "short_name", "type", "typespecs", "update_mode", "volume", "variable_tick_size", "value_unit_id"}},
	}

	for _, token := range authMsgs {
		if err := tv_api.sendToWriteChannel(token.name, token.args); err != nil {
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

func (tv_api *TV_API) readMessage(buffer string) error {
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

			var res map[string]interface{} = make(map[string]interface{})
			res["symbol"] = info["n"]
			if data, ok := info["v"].(map[string]interface{}); ok {
				res["volume"] = data["volume"]
				res["current_price"] = data["lp"]
				res["price_change"] = data["ch"]
				res["price_change_percentage"] = data["chp"]
				res["timestamp"] = data["lp_time"]
			}

			tv_api.ReadCh <- res
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
			res["symbol"] = tv_api.seriesMap[seriesId]
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

			tv_api.ReadCh <- res
		} else if tv_api.halts.haltedOn != "" && res["m"] == tv_api.halts.haltedOn {
			tv_api.halts.mu.Unlock()
			tv_api.halts.haltedOn = ""
		} else if res["m"] == "critical_error" {
			return errors.New("readMessage: TradingView Critical Error: " + msg)
		} else if res["m"] == "protocol_error" {
			return errors.New("readMessage: TradingView Protocol Error: " + msg)
		}
	}
	return nil
}