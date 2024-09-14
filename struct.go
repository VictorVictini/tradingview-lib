package tradingview

import (
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

	halted halted
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
	api.halted = halted{
		requiredResponses: map[string]string{
			"create_series":     "series_completed",
			"modify_series":     "series_completed",
			"request_more_data": "series_completed",
			"resolve_symbol":    "symbol_resolved",
			"switch_timezone":   "tickmark_update",
		},
		on: "",
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
	symbols_conv := convertInterfaceArr(symbols)

	// sending data we want to the server
	err := api.sendWriteThread("quote_add_symbols", append([]interface{}{api.qssq}, symbols_conv...))
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
Updates what real time stocks/symbols are being provided by the server
*/
func (api *API) quoteFastSymbols() error {
	// retrieve keys then convert the slice to []interface{}
	symbols := slices.Collect(maps.Keys(api.realtimeSymbols))
	symbols_conv := convertInterfaceArr(symbols)

	// send the request to the server
	return api.sendWriteThread("quote_fast_symbols", append([]interface{}{api.qs}, symbols_conv...))
}
