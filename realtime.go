package tradingview

import (
	"maps"
	"slices"
)

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
	return api.updateRealtimeSymbols()
}

func (api *API) RemoveRealtimeSymbols(symbols []string) error {
	symbols_conv := convertInterfaceArr(symbols)
	if err := api.sendWriteThread("quote_remove_symbols", append([]interface{}{api.session.quote.symbolQuotes}, symbols_conv...)); err != nil {
		return err
	}

	for _, symbol := range symbols {
		delete(api.symbols.resolvedIDs, symbol)
	}

	return api.updateRealtimeSymbols()
}

/*
Updates what real time stocks/symbols are being provided by the server
*/
func (api *API) updateRealtimeSymbols() error {
	// retrieve keys then convert the slice to []interface{}
	symbols := slices.Collect(maps.Keys(api.symbols.realtimeSet))
	symbols_conv := convertInterfaceArr(symbols)

	// send the request to the server
	return api.sendWriteThread("quote_fast_symbols", append([]interface{}{api.session.quote.key}, symbols_conv...))
}
