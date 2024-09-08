package main

import (
	"fmt"
)

func main() {
	err := OpenConnection()
	if err != nil {
		fmt.Println("Error: ", err)
	} else {
		//
		//GetHistory("FOREXCOM:GBPJPY", "240")
		//GetHistory("BATS:LLY", "240")

		fmt.Println("getting history:")
		//AddRealtimeSymbols([]string{"FOREXCOM:GBPJPY"})
		GetHistory("FOREXCOM:GBPJPY", "240")

		fmt.Println("getting more candles:")
		RequestMoreData(5)
		//GetHistory("FOREXCOM:GBPUSD", "240")

		//time.Sleep(2 * time.Second)
		//GetHistory("FOREXCOM:GBPJPY", "240")
		fmt.Println("finish")
		select {} // block program
	}
}
