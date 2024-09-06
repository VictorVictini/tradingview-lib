package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

func OpenConnection() {
	url := "wss://data.tradingview.com/socket.io/websocket"
	header := http.Header{}
	header.Add("Origin", "https://www.tradingview.com")

	ws, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		fmt.Println("Dial error: ", err) // TODO quit program? return error?
		return
	}

	// separate thread for the msgs
	go func() {
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				fmt.Println("Read error: ", err)
				return
			}

			fmt.Println(readMessage(ws, string(message))) // TODO
		}
	}()

	// auth(ws)
}

func readMessage(ws *websocket.Conn, data string) string {
	// TODO
	return data
}

func main() {
	OpenConnection()
	select {} // block program
}
