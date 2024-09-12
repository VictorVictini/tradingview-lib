package main

import (
	"math/rand"
)

var csToken string = "cs_" + createToken()
var qsToken string = createToken()
var qs string = "qs_" + qsToken
var qssq string = "qs_snapshoter_basic-symbol-quotes_" + qsToken

var realtimeSymbols map[string]bool = make(map[string]bool) // goofy hashset

func createToken() string {
	token := make([]byte, TOKEN_LENGTH)

	for i := range token {
		randomIndex := rand.Intn(len(TOKEN_CHARS))
		token[i] = TOKEN_CHARS[randomIndex]
	}

	return string(token)
}
