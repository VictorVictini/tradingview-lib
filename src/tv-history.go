package main

import (
	"strconv"
)

var symbolCounter uint64 = 0
var seriesCounter uint64 = 0
var seriesCreated bool = false

var initHistoryCandles int = 10 // amount of candles to load at the start, then RequestMoreData can load more

var resolvedSymbols map[string]string = make(map[string]string)
var seriesMap map[string]string = make(map[string]string)

func resolveSymbol(symbol string, sessionType SessionType) error {
	if _, exists := resolvedSymbols[symbol]; exists {
		return nil
	}

	symbolCounter++
	id := "symbol_" + strconv.FormatUint(symbolCounter, 10) //symbol id

	err := sendMessage("resolve_symbol", []interface{}{csToken, id, "={\"symbol\":\"" + symbol + "\",\"adjustment\":\"splits\",\"session\":\"" + string(sessionType) + "\"}"})
	if err != nil {
		return err
	}

	resolvedSymbols[symbol] = id
	return nil
}
