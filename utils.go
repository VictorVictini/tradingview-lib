package tradingview

import (
	"errors"
	"math/rand"
)

/*
Converts a string array to an interface array
*/
func convertInterfaceArr(arr []string) []interface{} {
	inter := make([]interface{}, len(arr))
	for i := range arr {
		inter[i] = arr[i]
	}
	return inter
}

/*
Handles sending data to the write channel
and retrieves an error if one occurred (otherwise nil)
*/
func (api *API) sendWriteThread(name string, args []interface{}) error {
	// send data to the write channel
	api.Channels.write <- request{name, args}

	// retrieve any error that has occurred
	err, ok := <-api.Channels.internalError
	if !ok {
		return errors.New("sendWriteThread: internal error channel is closed")
	}
	return err
}

/*
Creates a token (a randomised string of characters)
*/
func createToken() string {
	token := make([]byte, TOKEN_LENGTH)
	for i := range token {
		randomIndex := rand.Intn(len(TOKEN_CHARS))
		token[i] = TOKEN_CHARS[randomIndex]
	}
	return string(token)
}
