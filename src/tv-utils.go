package tv_api

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
func (tv_api *TV_API) sendToWriteChannel(name string, args []interface{}) error {
	// send data to the write channel
	tv_api.writeCh <- map[string]interface{}{
		"name": name,
		"args": args,
	}

	// retrieve any error that has occurred
	err, ok := <-tv_api.internalErrorCh
	if !ok {
		return errors.New("sendToWriteChannel: internal error channel is closed")
	}
	return err
}
