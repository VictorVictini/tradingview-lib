package main

import (
	"fmt"
)

func main() {
	err := OpenConnection()
	if err != nil {
		fmt.Println("Error: ", err)
	} else {
		AddRealtimeSymbols([]string{"FOREXCOM:GBPJPY"})
		select {} // block program
	}
}
