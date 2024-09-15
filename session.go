package tradingview

import (
	"net/http"

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
			key: "cs_" + createToken(),
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

func (api *API) auth() error {
	authMsgs := []request{
		{"set_auth_token", []interface{}{"unauthorized_user_token"}},
		{"chart_create_session", []interface{}{api.session.chart.key, ""}},
		{"quote_create_session", []interface{}{api.session.quote.key}},
		{"quote_create_session", []interface{}{api.session.quote.symbolQuotes}},
		{"quote_set_fields", []interface{}{api.session.quote.symbolQuotes, "base-currency-logoid", "ch", "chp", "currency-logoid", "currency_code", "currency_id", "base_currency_id", "current_session", "description", "exchange", "format", "fractional", "is_tradable", "language", "local_description", "listed_exchange", "logoid", "lp", "lp_time", "minmov", "minmove2", "original_name", "pricescale", "pro_name", "short_name", "type", "typespecs", "update_mode", "volume", "variable_tick_size", "value_unit_id"}},
	}

	for _, token := range authMsgs {
		if err := api.sendWriteThread(token.name, token.args); err != nil {
			return err
		}
	}

	return nil
}