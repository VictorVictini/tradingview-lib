package main

import (
	"fmt"
	"log"
)

func main() {
	var tv_api TV_API
	err := tv_api.OpenConnection()
	if err != nil {
		fmt.Println("Error: ", err)
	} else {
		var timeframe Timeframe = OneMinute
		err := tv_api.GetHistory("FOREXCOM:GBPJPY", timeframe, "regular")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("DONESO")
		err = tv_api.GetHistory("BATS:LLY", timeframe, "regular")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("getting history:")

		err = tv_api.AddRealtimeSymbols([]string{"FOREXCOM:EURJPY"})
		if err != nil {
			log.Fatal(err)
		}
		//GetHistory("FOREXCOM:GBPJPY", timeframe, "regular")
		//AddRealtimeSymbols([]string{"FOREXCOM:GBPJPY"})
		//GetHistory("BATS:LLY", timeframe, "regular")
		//SwitchTimezone("Europe/London")

		//fmt.Println("getting more candles:")
		//RequestMoreData(50)
		//GetHistory("FOREXCOM:GBPUSD", timeframe, "regular")

		//time.Sleep(2 * time.Second)
		//GetHistory("FOREXCOM:GBPJPY", timeframe, "regular")
		//RemoveRealtimeSymbols([]string{"FOREXCOM:EURJPY"})
		//fmt.Println("finish")
		select {} // block program
	}
}
