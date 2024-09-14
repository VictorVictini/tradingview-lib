package tradingview

import (
	"errors"
	"math/rand"
)

func convertStringArrToInterfaceArr(strArr []string) []interface{} {
	inter := make([]interface{}, len(strArr))
	for i := range strArr {
		inter[i] = strArr[i]
	}

	return inter
}

/*
Handles sending data to the write channel
and retrieves an error if one occurred (otherwise nil)
*/
func (api *API) sendWriteThread(name string, args []interface{}) error {
	// send data to the write channel
	api.writeCh <- map[string]interface{}{
		"name": name,
		"args": args,
	}

	// retrieve any error that has occurred
	err, ok := <-api.internalErrorCh
	if !ok {
		return errors.New("sendWriteThread: internal error channel is closed")
	}
	return err
}

func createToken() string {
	token := make([]byte, TOKEN_LENGTH)

	for i := range token {
		randomIndex := rand.Intn(len(TOKEN_CHARS))
		token[i] = TOKEN_CHARS[randomIndex]
	}

	return string(token)
}
