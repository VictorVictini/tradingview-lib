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

		// unlock the mutex if the requested string has been returned by the server
		if api.halted.on != "" && res["m"] == api.halted.on {
			api.halted.mutex.Unlock()
			api.halted.on = "" // resetting the mutex

			// otherwise handle the response
		} else {
			api.handler(res, msg)
		}
	}
	return nil
}

func (api *API) handler(data map[string]interface{}, msg string) error {
	switch data["m"] {

	// if the server provided realtime price changes
	case "qsd":
		// ensure the data is in a valid format
		resp, ok := data["p"].([]interface{})
		if !ok {
			return nil
		}
		info, ok := resp[1].(map[string]interface{})
		if !ok {
			return nil
		}

		// make the data more readable
		var result map[string]interface{} = make(map[string]interface{})
		result["symbol"] = info["n"]
		if data, ok := info["v"].(map[string]interface{}); ok {
			result["volume"] = data["volume"]
			result["current_price"] = data["lp"]
			result["price_change"] = data["ch"]
			result["price_change_percentage"] = data["chp"]
			result["timestamp"] = data["lp_time"]
		}

		// send to the read thread for the user to use
		api.Channels.Read <- result

		// get historical data
	case "timescale_update":

		// ensure the data is in a valid format
		resp, ok := data["p"].([]interface{})
		if !ok {
			return nil
		}
		info, ok := resp[1].(map[string]interface{})
		if !ok {
			return nil
		}
		seriesInfo, ok := info[HISTORY_TOKEN].(map[string]interface{})
		if !ok {
			return nil
		}
		allData, ok := seriesInfo["s"].([]interface{})
		if !ok {
			return nil
		}
		seriesId, ok := seriesInfo["t"].(string)
		if !ok {
			return nil
		}

		// more readable structure to store data in
		var result map[string]interface{} = make(map[string]interface{})
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
		result["symbol"] = api.series.mapsSymbols[seriesId]
		result["timestamp"] = timestamp
		result["open"] = open
		result["high"] = high
		result["low"] = low
		result["close"] = close
		result["volume"] = volume

		// provide the data to the read channel for the user to receive
		api.Channels.Read <- result

		// return an error if the server had returned either error
	case "critical_error":
		return errors.New("parseServerMessage: TradingView Critical Error: " + msg)
	case "protocol_error":
		return errors.New("parseServerMessage: TradingView Protocol Error: " + msg)

	}
	return nil
}
