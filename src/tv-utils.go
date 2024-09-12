package main

func convertStringArrToInterfaceArr(strArr []string) []interface{} {
	inter := make([]interface{}, len(strArr))
	for i := range strArr {
		inter[i] = strArr[i]
	}

	return inter
}
