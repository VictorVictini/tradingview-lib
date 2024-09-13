package tradingview

import (
	"math/rand"
)

func createToken() string {
	token := make([]byte, TOKEN_LENGTH)

	for i := range token {
		randomIndex := rand.Intn(len(TOKEN_CHARS))
		token[i] = TOKEN_CHARS[randomIndex]
	}

	return string(token)
}
