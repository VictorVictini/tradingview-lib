package tradingview

import (
	"strconv"
)

func (tv_api *TV_API) resolveSymbol(symbol string, sessionType SessionType) error {
	if _, exists := tv_api.resolvedSymbols[symbol]; exists {
		return nil
	}

	tv_api.symbolCounter++
	id := "symbol_" + strconv.FormatUint(tv_api.symbolCounter, 10) //symbol id

	err := tv_api.sendToWriteChannel("resolve_symbol", []interface{}{tv_api.csToken, id, "={\"symbol\":\"" + symbol + "\",\"adjustment\":\"splits\",\"session\":\"" + string(sessionType) + "\"}"})
	if err != nil {
		return err
	}

	tv_api.resolvedSymbols[symbol] = id
	return nil
}
