package tradingview

import (
	"strconv"
)

func (api *API) GetHistory(symbol string, timeframe Timeframe, sessionType SessionType) error {
	err := api.resolveSymbol(symbol, sessionType)
	if err != nil {
		return err
	}

	api.seriesCounter++
	series := "s" + strconv.FormatUint(api.seriesCounter, 10)
	id := api.resolvedSymbols[symbol]

	api.seriesMap[series] = symbol

	// possibly use sync.Once?
	if !api.seriesCreated {
		api.seriesCreated = true
		return api.sendWriteThread("create_series", []interface{}{api.csToken, HISTORY_TOKEN, series, id, string(timeframe), INITIAL_HISTORY_CANDLES, ""})
	}
	return api.sendWriteThread("modify_series", []interface{}{api.csToken, HISTORY_TOKEN, series, id, string(timeframe), ""})
}

func (api *API) resolveSymbol(symbol string, sessionType SessionType) error {
	if _, exists := api.resolvedSymbols[symbol]; exists {
		return nil
	}

	api.symbolCounter++
	id := "symbol_" + strconv.FormatUint(api.symbolCounter, 10) //symbol id

	err := api.sendWriteThread("resolve_symbol", []interface{}{api.csToken, id, "={\"symbol\":\"" + symbol + "\",\"adjustment\":\"splits\",\"session\":\"" + string(sessionType) + "\"}"})
	if err != nil {
		return err
	}

	api.resolvedSymbols[symbol] = id
	return nil
}
