package tradingview

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

/*
Switches the timezone the data is viewed on for the current session
*/
func (api *API) SwitchTimezone(timezone string) error {
	return api.sendWriteThread("switch_timezone", append([]interface{}{api.session.chart.key}, timezone))
}

/*
Sends the data to the server in the format it requests
*/
func (api *API) sendServerMessage(name string, args []interface{}) error {
	// change input into requested format
	message, err := json.Marshal(
		map[string]interface{}{
			"m": name,
			"p": args,
		},
	)

	// ensure the data was created without issues
	if err != nil {
		return err
	}

	// send the message to the server
	return api.ws.WriteMessage(websocket.TextMessage, []byte(SEPARATOR+strconv.Itoa(len(message))+SEPARATOR+string(message)))
}

/*
Parse the message from the server
*/
func (api *API) parseServerMessage(buffer string) error {
	// separate each message by the separator
	msgs := strings.Split(buffer, SEPARATOR)

	// for each message
	for _, msg := range msgs {
		// parse the message through JSON
		var res map[string]interface{}
		err := json.Unmarshal([]byte(msg), &res)

		// the message was not JSON
		if err != nil {
			// send the message back if it contains the identifier
			if strings.Contains(msg, "~h~") {
				if err = api.ws.WriteMessage(websocket.TextMessage, []byte(SEPARATOR+strconv.Itoa(len(msg))+SEPARATOR+msg)); err != nil {
					return err
				}
			}

			continue
		}

		// the message was valid JSON
		// if the server provided realtime price changes
		if res["m"] == "qsd" {
			// ensure the data is in a valid format
			resp, ok := res["p"].([]interface{})
			if !ok {
				continue
			}
			info, ok := resp[1].(map[string]interface{})
			if !ok {
				continue
			}

			// make the data more readable
			var res map[string]interface{} = make(map[string]interface{})
			res["symbol"] = info["n"]
			if data, ok := info["v"].(map[string]interface{}); ok {
				res["volume"] = data["volume"]
				res["current_price"] = data["lp"]
				res["price_change"] = data["ch"]
				res["price_change_percentage"] = data["chp"]
				res["timestamp"] = data["lp_time"]
			}

			// send to the read thread for the user to use
			api.Channels.Read <- res

			// get historical data
		} else if res["m"] == "timescale_update" {
			// ensure the data is in a valid format
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

			// more readable structure to store data in
			var res map[string]interface{} = make(map[string]interface{})
			var timestamp, open, high, low, close, volume []interface{}

			// for all the data provided
			for _, dataElement := range allData {
				// ensure it is in a valid format
				dataElement, ok := dataElement.(map[string]interface{})
				if !ok {
					continue
				}
				data, ok := dataElement["v"].([]interface{})
				if !ok {
					continue
				}

				// add to the parallel arrays
				timestamp = append(timestamp, data[0])
				open = append(open, data[1])
				high = append(high, data[2])
				low = append(low, data[3])
				close = append(close, data[4])

				// add the volume as nil if we can't add it
				if len(data) >= 6 {
					volume = append(volume, data[5])
				} else {
					volume = append(volume, nil)
				}
			}

			// move all the data into the usable data structure
			res["symbol"] = api.series.mapsSymbols[seriesId]
			res["timestamp"] = timestamp
			res["open"] = open
			res["high"] = high
			res["low"] = low
			res["close"] = close
			res["volume"] = volume

			// provide the data to the read channel for the user to receive
			api.Channels.Read <- res

			// unlock the mutex if the requested string has been returned by the server
		} else if api.halted.on != "" && res["m"] == api.halted.on {
			api.halted.mutex.Unlock()
			api.halted.on = "" // resetting the mutex

			// return an error if the server had returned either error
		} else if res["m"] == "critical_error" {
			return errors.New("parseServerMessage: TradingView Critical Error: " + msg)
		} else if res["m"] == "protocol_error" {
			return errors.New("parseServerMessage: TradingView Protocol Error: " + msg)
		}
	}
	return nil
}
