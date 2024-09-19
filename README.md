# TradingView Library
- tradingview-lib is a Go based library that sends and receives messages from TradingView using websockets. To fully understand and utilise this project, we suggest understanding how TradingView receives and sends data via its websockets (specifically what it is sending/receiving)

## Installation
- `go get github.com/VictorVictini/tradingview-lib`
- import `github.com/VictorVictini/tradingview-lib`

## Features
- Open a connection to TradingView to send/receive messages such as:
  - Receive stock data
    - Symbol, High, Low, Open, Close, Volume and seconds elapsed from Jan 1st 1970 (Unix time standard)
  - Add/Remove real-time tracking of stock changes 
    - Will output the most recent change to a stock. Will output `<nil>` for entries that have no change.
    - Choose the stock via `AddRealtimeSymbols()`, `RemoveRealtimeSymbols()` to stop receiving data about the stock.
  - Get historical stock data from a specific point in time.
    - You can decide the stock, date, timeframe between each candle and session type using the `GetHistory()` function.
    - To receive further historical data, you can invoke `RequestMoreCandles()` with a specific count to view even more data for the current `GetHistory()` options.
  - Change Timezone
    - `SwitchTimezone()` allows accurate assignment of stock data to specific times.
  - Error handling for most functions that handle data transfer between TradingView and this library
    - Keep in mind the program will perform unexpectedly if an invalid input is received into these functions.
