package main

import (
	"fmt"
)

func main() {
	err := OpenConnection()
	if err != nil {
		fmt.Println("Error: ", err)
	} else {
		//AddRealtimeSymbols([]string{"FOREXCOM:GBPJPY"})
		GetHistory("FOREXCOM:GBPJPY", "240")
		//GetHistory("FOREXCOM:GBPUSD", "240")
		select {} // block program
	}
}
