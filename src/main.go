package tv_lib

import (
	"fmt"
)

func main() {
	err := OpenConnection()
	if err != nil {
		fmt.Println("Error: ", err)
	} else {
		//
		var timeframe Timeframe = OneMinute
		GetHistory("FOREXCOM:GBPJPY", timeframe, "regular")
		//GetHistory("BATS:LLY", timeframe, "regular")

		//fmt.Println("getting history:")

		//AddRealtimeSymbols([]string{"FOREXCOM:EURJPY"})
		//AddRealtimeSymbols([]string{"FOREXCOM:EURJPY"})
		//AddRealtimeSymbols([]string{"FOREXCOM:EURJPY"})
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
