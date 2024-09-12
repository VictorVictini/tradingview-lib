package main

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

/*
Handles data associated with an instance of the websocket
*/
type TV_API struct {
	ws      *websocket.Conn
	readCh  chan map[string]interface{}
	writeCh chan map[string]interface{}
	errorCh chan error // receives errors that occurred in read/write threads
	halts   await_response
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
-- requires some thinking for error handling
*/
func (tv_api *TV_API) activeWriteListener() {
	for {
		data := <-tv_api.writeCh

		// ensure the name is valid
		if data["name"] == nil {
			tv_api.errorCh <- errors.New("activeWriteListener: \"name\" property not provided to the channel")
			continue
		}
		name, ok := data["name"].(string)
		if !ok {
			tv_api.errorCh <- errors.New("activeWriteListener: \"name\" property is not of type string")
			continue
		}

		// ensure the arguments are valid
		if data["args"] == nil {
			tv_api.errorCh <- errors.New("activeWriteListener: \"args\" property not provided to the channel")
			continue
		}
		args, ok := data["args"].([]interface{})
		if !ok {
			tv_api.errorCh <- errors.New("activeWriteListener: \"args\" property is not of type []interface{}")
			continue
		}

		// mutex handling later

		if err := tv_api.sendMessage(name, args); err != nil {
			tv_api.errorCh <- err
		}
	}
}
