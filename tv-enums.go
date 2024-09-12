package main

type SessionType string
type Timeframe string

const (
	Regular  SessionType = "regular"
	Extended SessionType = "extended"
)

const (
	OneMinute        Timeframe = "1"
	ThreeMinutes     Timeframe = "3"
	FiveMinutes      Timeframe = "5"
	FifteenMinutes   Timeframe = "15"
	ThirtyMinutes    Timeframe = "30"
	FortyFiveMinutes Timeframe = "45"

	OneHour    Timeframe = "60"
	TwoHours   Timeframe = "120"
	ThreeHours Timeframe = "180"
	FourHours  Timeframe = "240"

	OneDay       Timeframe = "1D"
	OneWeek      Timeframe = "1W"
	OneMonth     Timeframe = "1M"
	ThreeMonths  Timeframe = "3M"
	SixMonths    Timeframe = "6M"
	TwelveMonths Timeframe = "12M"
)
