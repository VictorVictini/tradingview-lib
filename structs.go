package tradingview

import (
	"sync"

	"github.com/gorilla/websocket"
)

/*
Handles data associated with an instance of the websocket
*/
type API struct {
	Channels Channels

	ws      *websocket.Conn
	series  series
	symbols symbols
	session session
	halted  halted
}

/*
Handles data transferring between threads
as well as the channels the user will be able to utilise
*/
type Channels struct {
	Read          chan map[string]interface{}
	write         chan request
	Error         chan error // receives errors that occurred in read/write threads
	internalError chan error // internal handling of errors in read/write threads
}

/*
Handles data related to series
*/
type series struct {
	counter               uint64
	wasCreated            bool
	initialHistoryCandles int               // how many candles to load at the start of GetHistory?
	mapsSymbols           map[string]string // maps a series to a correlating symbol
}

/*
Handles data related to symbols
*/
type symbols struct {
	counter     uint64
	resolvedIDs map[string]string // correlating IDs for the given symbols
	realtimeSet map[string]bool   // set of all currently active realtime symbols
}

/*
Handles session data
*/
type session struct {
	chart chart
	quote quote
}

/*
Handles chart session data
*/
type chart struct {
	key string
}

/*
Handles quote session data
*/
type quote struct {
	token        string
	key          string
	symbolQuotes string
}

/*
Handles waiting until a specific response is provided by the server
*/
type halted struct {
	mutex             sync.Mutex
	requiredResponses map[string]string // halts write thread until the correlating response is provided by the server
	on                string
}

/*
Handles the structure of which data sent to the server is formatted
*/
type request struct {
	name string
	args []interface{}
}
