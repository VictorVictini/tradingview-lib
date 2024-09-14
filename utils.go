package tradingview

import "errors"

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
func (api *API) sendToWriteChannel(name string, args []interface{}) error {
	// send data to the write channel
	api.writeCh <- map[string]interface{}{
		"name": name,
		"args": args,
	}

	// retrieve any error that has occurred
	err, ok := <-api.internalErrorCh
	if !ok {
		return errors.New("sendToWriteChannel: internal error channel is closed")
	}
	return err
}
