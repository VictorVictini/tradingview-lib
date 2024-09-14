package tradingview

import "errors"

/*
Active listener that receives data from the server
which is later parsed then passed to the read channel
*/
func (api *API) activeReadListener() {
	for {
		_, message, err := api.ws.ReadMessage()
		if err != nil {
			api.ErrorCh <- err
			return // quit reading if error
		}

		err = api.readMessage(string(message))
		if err != nil {
			api.ErrorCh <- err
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
		data, ok := <-api.writeCh
		if !ok {
			err := errors.New("activeWriteListener: write channel is closed")
			api.internalErrorCh <- err
			api.ErrorCh <- err
			return
		}

		// ensure the name is valid
		_, ok = data["name"]
		if !ok {
			err := errors.New("activeWriteListener: \"name\" property not provided to the channel")
			api.internalErrorCh <- err
			api.ErrorCh <- err
			continue
		}
		name, ok := data["name"].(string)
		if !ok {
			err := errors.New("activeWriteListener: \"name\" property is not of type string")
			api.internalErrorCh <- err
			api.ErrorCh <- err
			continue
		}

		// ensure the arguments are valid
		_, ok = data["args"]
		if !ok {
			err := errors.New("activeWriteListener: \"args\" property not provided to the channel")
			api.internalErrorCh <- err
			api.ErrorCh <- err
			continue
		}
		args, ok := data["args"].([]interface{})
		if !ok {
			err := errors.New("activeWriteListener: \"args\" property is not of type []interface{}")
			api.internalErrorCh <- err
			api.ErrorCh <- err
			continue
		}

		// handle errors from sending the data to the server
		err := api.sendMessage(name, args)
		if err != nil {
			api.ErrorCh <- err
		}
		api.internalErrorCh <- err

		// lock the write thread until a given response is received (if necessary)
		if haltedOn, ok := api.halts.requiredResponses[name]; ok {
			api.halts.haltedOn = haltedOn
			api.halts.mu.Lock()
		}
	}
}
