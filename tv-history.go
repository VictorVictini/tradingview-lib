package main

import (
	"strconv"
)

var symbolCounter int = 0
var seriesCounter int = 0
var seriesCreated bool = false

var maxHistoryCandles int = 10 // TODO make this changeable at start?

var resolvedSymbols map[string]string = make(map[string]string)
var seriesMap map[string]string = make(map[string]string)

func resolveSymbol(symbol string) error { // generalised params in sendMessage
    if _, exists := resolvedSymbols[symbol]; exists {
        return nil
    }

	symbolCounter++
    id := "symbol_" + strconv.Itoa(symbolCounter) //symbol id

    err := sendMessage("resolve_symbol", []interface{} {csToken, id, "={\"symbol\":\"" + symbol + "\",\"adjustment\":\"splits\",\"session\":\"regular\"}"});
    if err != nil {
        return err
    }

    resolvedSymbols[symbol] = id
    return nil
}
