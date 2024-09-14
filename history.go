package tradingview

import (
	"strconv"
)

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
