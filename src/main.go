package main

import (
	"fmt"
	"log"
)

func main() {
	var timeframe Timeframe = OneMinute
	// create the api variable
	var tv_api TV_API

	// open the connection
	err := tv_api.OpenConnection()
	if err != nil {
		log.Fatal("main: ", err)
	}

	// create an active receiver for the channels
	go ActiveReceiver(&tv_api)

	// random testing
	err = tv_api.GetHistory("FOREXCOM:GBPJPY", timeframe, "regular")
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

	fmt.Println("getting more candles:")
	tv_api.RequestMoreData(5)
	tv_api.RequestMoreData(5)
	tv_api.RequestMoreData(5)
	tv_api.RequestMoreData(5)
	tv_api.RequestMoreData(5)
	tv_api.RequestMoreData(5)
	tv_api.RequestMoreData(5)
	tv_api.RequestMoreData(5)
	//GetHistory("FOREXCOM:GBPUSD", timeframe, "regular")

	//time.Sleep(2 * time.Second)
	//GetHistory("FOREXCOM:GBPJPY", timeframe, "regular")
	//RemoveRealtimeSymbols([]string{"FOREXCOM:EURJPY"})
	//fmt.Println("finish")
	select {}
}

func ActiveReceiver(tv_api *TV_API) {
	fmt.Println("activeReceiver enabled")
	for {
		select {
		case data := <-tv_api.readCh:
			fmt.Println("received: ", data)
		case errMsg := <-tv_api.errorCh:
			fmt.Println(errMsg)
		}
	}
}
