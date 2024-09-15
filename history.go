package tradingview

import "strconv"

/*
Retrieves 10 candles from history for the provided symbol/stock
Provides candles for the requested timeframe
*/
func (api *API) GetHistory(symbol string, timeframe Timeframe, sessionType SessionType) error {
	// resolve the symbol
	err := api.resolveSymbol(symbol, sessionType)
	if err != nil {
		return err
	}

	// create the series and symbol IDs
	api.series.counter++
	seriesID := "s" + strconv.FormatUint(api.series.counter, 10)
	symbolID := api.symbols.resolvedIDs[symbol]

	// map the symbol to the related series
	api.series.mapsSymbols[seriesID] = symbol

	// for the first instance of GetHistory(), create the initial series
	if !api.series.wasCreated {
		api.series.wasCreated = true // to avoid repeating this if statement
		return api.sendWriteThread("create_series", []interface{}{api.session.chart.key, HISTORY_TOKEN, seriesID, symbolID, string(timeframe), INITIAL_HISTORY_CANDLES, ""})
	}

	// not the first instance, so modify the series instead
	return api.sendWriteThread("modify_series", []interface{}{api.session.chart.key, HISTORY_TOKEN, seriesID, symbolID, string(timeframe), ""})
}

/*
Retrieves more data of the most recently loaded symbol,
requires GetHistory() to have been used before it
*/
func (api *API) RequestMoreData(candleCount int) error {
	return api.sendWriteThread("request_more_data", append([]interface{}{api.session.chart.key}, HISTORY_TOKEN, candleCount))
}

/*
Adds the symbol to the set of resolved symbols if needed
*/
func (api *API) resolveSymbol(symbol string, sessionType SessionType) error {
	// symbol exists, so ignore it
	if _, exists := api.symbols.resolvedIDs[symbol]; exists {
		return nil
	}

	// create the symbol id
	api.symbols.counter++
	symbolID := "symbol_" + strconv.FormatUint(api.symbols.counter, 10)

	// send server the symbol to resolve
	err := api.sendWriteThread("resolve_symbol", []interface{}{api.session.chart.key, symbolID, "={\"symbol\":\"" + symbol + "\",\"adjustment\":\"splits\",\"session\":\"" + string(sessionType) + "\"}"})
	if err != nil {
		return err
	}

	// add to the set of resolved symbols
	api.symbols.resolvedIDs[symbol] = symbolID
	return nil
}
