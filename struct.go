package tradingview

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
type API struct {
	ws *websocket.Conn

	ReadCh          chan map[string]interface{}
	writeCh         chan map[string]interface{}
	ErrorCh         chan error // receives errors that occurred in read/write threads
	internalErrorCh chan error // internal handling of errors in read/write threads

	symbolCounter   uint64
	seriesCounter   uint64
	seriesCreated   bool // use sync.Once in clean up
	resolvedSymbols map[string]string
	seriesMap       map[string]string

	csToken string
	qsToken string
	qs      string
	qssq    string

	realtimeSymbols map[string]bool

	halts await_response
}

/*
Handles waiting until a specific response is provided by the server
*/
type await_response struct {
	mu                sync.Mutex
	requiredResponses map[string]string // halts write thread until the correlating response is provided by the server
	haltedOn          string
}

/*
Creates an active websocket connection
*/
func (api *API) OpenConnection() error {
	// setting up the header
	header := http.Header{}
	header.Add("Origin", TV_ORIGIN_URL)

	// creating the websocket
	ws, _, err := websocket.DefaultDialer.Dial(TV_URL, header)
	if err != nil {
		return err
	}

	// initialise in values for the struct
	api.ws = ws

	api.ReadCh = make(chan map[string]interface{})
	api.writeCh = make(chan map[string]interface{})
	api.ErrorCh = make(chan error)
	api.internalErrorCh = make(chan error)

	api.symbolCounter = 0
	api.seriesCounter = 0
	api.seriesCreated = false // use sync.Once in clean up
	api.resolvedSymbols = make(map[string]string)
	api.seriesMap = make(map[string]string)

	api.csToken = "cs_" + createToken()
	api.qsToken = createToken()
	api.qs = "qs_" + api.qsToken
	api.qssq = "qs_snapshoter_basic-symbol-quotes_" + api.qsToken
	api.realtimeSymbols = make(map[string]bool)

	// required responses for a given request being sent
	// halts write requests from being sent until it is received
	api.halts = await_response{
		requiredResponses: map[string]string{
			"create_series":     "series_completed",
			"modify_series":     "series_completed",
			"request_more_data": "series_completed",
			"resolve_symbol":    "symbol_resolved",
			"switch_timezone":   "tickmark_update",
		},
		haltedOn: "",
	}

	// thread to actively read messages from the websocket to a channel
	go api.activeReadListener()

	// thread to actively write messages from the channel to the websocket
	go api.activeWriteListener()

	// authenticate the websocket
	return api.auth()
}

/*
Retrieves real-time data for the given stocks/symbols
which is then provided to the read channel
*/
func (api *API) AddRealtimeSymbols(symbols []string) error {
	// converts symbols from []string to []interface{}
	symbols_conv := convertStringArrToInterfaceArr(symbols)

	// sending data we want to the server
	err := api.sendToWriteChannel("quote_add_symbols", append([]interface{}{api.qssq}, symbols_conv...))
	if err != nil {
		return err
	}

	// add the symbols to the set of handled symbols
	for _, symbol := range symbols {
		api.realtimeSymbols[symbol] = true
	}

	// tells server to start sending the symbols' real time data
	return api.quoteFastSymbols()
}

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

/*
Updates what real time stocks/symbols are being provided by the server
*/
func (api *API) quoteFastSymbols() error {
	// retrieve keys then convert the slice to []interface{}
	symbols := slices.Collect(maps.Keys(api.realtimeSymbols))
	symbols_conv := convertStringArrToInterfaceArr(symbols)

	// send the request to the server
	return api.sendToWriteChannel("quote_fast_symbols", append([]interface{}{api.qs}, symbols_conv...))
}
