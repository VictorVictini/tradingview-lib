package tradingview

import (
	"maps"
	"net/http"
	"slices"

	"github.com/gorilla/websocket"
)

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

	// manages channels for the user to access as well as internal channels
	api.Channels = Channels{
		Read:          make(chan map[string]interface{}),
		write:         make(chan request),
		Error:         make(chan error),
		internalError: make(chan error),
	}

	api.series = series{
		counter:     0,
		wasCreated:  false,
		mapsSymbols: make(map[string]string),
	}

	api.symbols = symbols{
		counter:     0,
		resolvedIDs: make(map[string]string),
		realtimeSet: make(map[string]bool),
	}

	api.session = session{
		chart: chart{
			token: "cs_" + createToken(),
		},
		quote: quote{
			token:        createToken(),
			key:          "qs_" + api.session.quote.token,
			symbolQuotes: "qs_snapshoter_basic-symbol-quotes_" + api.session.quote.token,
		},
	}

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
	err := api.sendWriteThread("quote_add_symbols", append([]interface{}{api.session.quote.symbolQuotes}, symbols_conv...))
	if err != nil {
		return err
	}

	// add the symbols to the set of handled symbols
	for _, symbol := range symbols {
		api.symbols.realtimeSet[symbol] = true
	}

	// tells server to start sending the symbols' real time data
	return api.quoteFastSymbols()
}

/*
Updates what real time stocks/symbols are being provided by the server
*/
func (api *API) quoteFastSymbols() error {
	// retrieve keys then convert the slice to []interface{}
	symbols := slices.Collect(maps.Keys(api.symbols.realtimeSet))
	symbols_conv := convertInterfaceArr(symbols)

	// send the request to the server
	return api.sendWriteThread("quote_fast_symbols", append([]interface{}{api.session.quote.key}, symbols_conv...))
}
