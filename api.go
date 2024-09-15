package tradingview

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

func (api *API) SwitchTimezone(timezone string) error {
	return api.sendWriteThread("switch_timezone", append([]interface{}{api.session.chart.key}, timezone))
}

func (api *API) sendServerMessage(name string, args []interface{}) error {
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
	msgs := strings.Split(buffer, SEPARATOR)
	for _, msg := range msgs {
		var res map[string]interface{}
		err := json.Unmarshal([]byte(msg), &res)

		// not a json string
		if err != nil {
			// not json
			if strings.Contains(msg, "~h~") {
				if err = api.ws.WriteMessage(websocket.TextMessage, []byte(SEPARATOR+strconv.Itoa(len(msg))+SEPARATOR+msg)); err != nil {
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
