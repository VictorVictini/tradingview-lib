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
		//GetHistory("FOREXCOM:GBPJPY", "240", "regular")
		//GetHistory("BATS:LLY", "240", "regular")

		//fmt.Println("getting history:")
		//AddRealtimeSymbols([]string{"FOREXCOM:GBPJPY"})
		//AddRealtimeSymbols([]string{"FOREXCOM:EURJPY"})
		//GetHistory("FOREXCOM:GBPJPY", "240", "regular")
		//SwitchTimezone("Europe/London")

		//fmt.Println("getting more candles:")
		//RequestMoreData(5)
		//GetHistory("FOREXCOM:GBPUSD", "240", "regular")

		//time.Sleep(2 * time.Second)
		//GetHistory("FOREXCOM:GBPJPY", "240", "regular")
		//RemoveRealtimeSymbols([]string{"FOREXCOM:EURJPY"})
		//fmt.Println("finish")
		select {} // block program
	}
}