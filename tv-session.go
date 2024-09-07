package main

import (
	"math/rand"
)

const TOKEN_LENGTH int = 12

var csToken string = "cs_" + createToken()
var qsToken string = createToken()
var qs string = "qs_" + qsToken
var qssq string = "qs_snapshoter_basic-symbol-quotes_" + qsToken

func createToken() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	token := make([]byte, TOKEN_LENGTH)

	for i := range token {
		randomIndex := rand.Intn(len(chars))
		token[i] = chars[randomIndex]
	}

	return string(token)
}
