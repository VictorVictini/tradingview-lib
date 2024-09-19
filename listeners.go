package tradingview

import "errors"

/*
Active listener that receives data from the server
which is later parsed then passed to the read channel
*/
func (api *API) activeReadListener() {
	for {
		// read the message from the server
		_, message, err := api.ws.ReadMessage()
		if err != nil {
			api.Channels.Error <- err
			return // quit reading if error
		}

		// parse the message
		err = api.parseServerMessage(string(message))
		if err != nil {
			api.Channels.Error <- err
			return
		}
	}
}

/*
Active listener that receives data from the writeCh channel
which is later sent to the server at the next available instance
*/
func (api *API) activeWriteListener() {
	for {
		// retrieve data from the write channel
		data, ok := <-api.Channels.write

		// error if it is closed
		if !ok {
			err := errors.New("activeWriteListener: write channel is closed")
			api.Channels.internalError <- err
			api.Channels.Error <- err
			return
		}

		// handle errors from sending the data to the server
		err := api.sendServerMessage(data.name, data.args)
		if err != nil {
			api.Channels.Error <- err
		}

		// always send so the function using internalError doesn't end up locked
		api.Channels.internalError <- err

		// lock the write thread until a given response is received (if necessary)
		if haltedOn, ok := api.halted.requiredResponses[data.name]; ok {
			api.halted.on = haltedOn
			api.halted.mutex.Lock()
		}
	}
}
