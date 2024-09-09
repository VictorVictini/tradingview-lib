package main

import (
	"strconv"
)

var symbolCounter int = 0
var seriesCounter int = 0
var seriesCreated bool = false

var initHistoryCandles int = 10 // amount of candles to load at the start, then RequestMoreData can load more

var resolvedSymbols map[string]string = make(map[string]string)
var seriesMap map[string]string = make(map[string]string)

func resolveSymbol(symbol string, sessionType string) error { // session type can be either "regular" or "extended"
	if _, exists := resolvedSymbols[symbol]; exists {
		return nil
	}

	symbolCounter++
	id := "symbol_" + strconv.Itoa(symbolCounter) //symbol id

	err := sendMessage("resolve_symbol", []interface{}{csToken, id, "={\"symbol\":\"" + symbol + "\",\"adjustment\":\"splits\",\"session\":\"" + sessionType + "\"}"})
	if err != nil {
		return err
	}

	resolvedSymbols[symbol] = id
	return nil
}
