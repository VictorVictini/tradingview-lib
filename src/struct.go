package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type TV_API struct {
	ws      *websocket.Conn
	readCh  chan []interface{}
	writeCh chan []interface{}
	errorCh chan error
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
	tv_api.readCh = make(chan []interface{})
	tv_api.writeCh = make(chan []interface{})
	tv_api.errorCh = make(chan error)

	// separate thread to actively read messages from the websocket
	go tv_api.activeListener()

	// authenticate the websocket
	return tv_api.auth()
}

// return to this later
func (tv_api *TV_API) activeListener() {
	for {
		_, message, err := tv_api.ws.ReadMessage()
		if err != nil {
			fmt.Println("Read error: ", err)
			return // quit reading if error
		}

		err = tv_api.readMessage(string(message))
		if err != nil {
			fmt.Println("OpenConnection: ", err)
			return
		}
	}
}
