package main

import (
	"errors"
	"maps"
	"net/http"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
)

/*
Handles data associated with an instance of the websocket
*/
type TV_API struct {
	ws              *websocket.Conn
	readCh          chan map[string]interface{}
	writeCh         chan map[string]interface{}
	errorCh         chan error // receives errors that occurred in read/write threads
	internalErrorCh chan error // internal handling of errors in read/write threads
	halts           await_response
}

/*
Handles waiting until a specific response is provided by the server
*/
type await_response struct {
	mu       sync.Mutex
	haltedOn string
}

/*
Creates an active websocket connection
*/
func (tv_api *TV_API) OpenConnection() error {
	// setting up the header
	header := http.Header{}
	header.Add("Origin", TV_ORIGIN_URL)

	// creating the websocket
	ws, _, err := websocket.DefaultDialer.Dial(TV_URL, header)
	if err != nil {
		return err
	}

	// fill in values for the struct
	tv_api.ws = ws
	tv_api.readCh = make(chan map[string]interface{})
	tv_api.writeCh = make(chan map[string]interface{})
	tv_api.errorCh = make(chan error)
	tv_api.internalErrorCh = make(chan error)

	// thread to actively read messages from the websocket to a channel
	go tv_api.activeReadListener()

	// thread to actively write messages from the channel to the websocket
	go tv_api.activeWriteListener()

	// authenticate the websocket
	return tv_api.auth()
}

/*
Active listener that receives data from the server
which is later parsed then passed to the readCh channel
*/
func (tv_api *TV_API) activeReadListener() {
	for {
		_, message, err := tv_api.ws.ReadMessage()
		if err != nil {
			tv_api.errorCh <- err
			return // quit reading if error
		}

		err = tv_api.readMessage(string(message))
		if err != nil {
			tv_api.errorCh <- err
			return
		}
	}
}

/*
Active listener that receives data from the writeCh channel
which is later sent to the server at the next available instance
*/
func (tv_api *TV_API) activeWriteListener() {
	for {
		data := <-tv_api.writeCh

		// ensure the name is valid
		_, ok := data["name"]
		if !ok {
			err := errors.New("activeWriteListener: \"name\" property not provided to the channel")
			tv_api.internalErrorCh <- err
			tv_api.errorCh <- err
			continue
		}
		name, ok := data["name"].(string)
		if !ok {
			err := errors.New("activeWriteListener: \"name\" property is not of type string")
			tv_api.internalErrorCh <- err
			tv_api.errorCh <- err
			continue
		}

		// ensure the arguments are valid
		_, ok = data["args"]
		if !ok {
			err := errors.New("activeWriteListener: \"args\" property not provided to the channel")
			tv_api.internalErrorCh <- err
			tv_api.errorCh <- err
			continue
		}
		args, ok := data["args"].([]interface{})
		if !ok {
			err := errors.New("activeWriteListener: \"args\" property is not of type []interface{}")
			tv_api.internalErrorCh <- err
			tv_api.errorCh <- err
			continue
		}

		// mutex handling later

		err := tv_api.sendMessage(name, args)
		if err != nil {
			tv_api.errorCh <- err
		}
		tv_api.internalErrorCh <- err
	}
}

/*
Retrieves real-time data for the given stocks/symbols
which is then provided to the read channel
*/
func (tv_api *TV_API) AddRealtimeSymbols(symbols []string) error {
	// converts symbols from []string to []interface{}
	symbols_conv := convertStringArrToInterfaceArr(symbols)

	// sending data we want to the server
	err := tv_api.sendToWriteChannel("quote_add_symbols", append([]interface{}{qssq}, symbols_conv...))
	if err != nil {
		return err
	}

	// add the symbols to the set of handled symbols
	for _, symbol := range symbols {
		realtimeSymbols[symbol] = true
	}

	// tells server to start sending the symbols' real time data
	return tv_api.quoteFastSymbols()
}

/*
Updates what real time stocks/symbols are being provided by the server
*/
func (tv_api *TV_API) quoteFastSymbols() error {
	// retrieve keys then convert the slice to []interface{}
	symbols := slices.Collect(maps.Keys(realtimeSymbols))
	symbols_conv := convertStringArrToInterfaceArr(symbols)

	// send the request to the server
	return tv_api.sendToWriteChannel("quote_fast_symbols", append([]interface{}{qs}, symbols_conv...))
}
