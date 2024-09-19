package tradingview

type SessionType string
type Timeframe string
type SearchType string

// session data
const (
	Regular  SessionType = "regular"
	Extended SessionType = "extended"
)

// timeframe data
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

// search data
const (
	All     SearchType = "undefined"
	Stocks  SearchType = "stocks"
	Forex   SearchType = "forex"
	Crypto  SearchType = "crypto"
	Indices SearchType = "index"
	Bonds   SearchType = "bond"
	Economy SearchType = "economic"
	Options SearchType = "options"
)

// URL data for initialisation
const TV_URL = "wss://data.tradingview.com/socket.io/websocket"
const TV_ORIGIN_URL = "https://www.tradingview.com"

// constants for token creation
const TOKEN_LENGTH = 12
const TOKEN_CHARS = "abcdefghijklmnopqrstuvwxyz0123456789"

// random constants
const SEPARATOR = "~m~"
const HISTORY_TOKEN = "sds_1"
const SEARCH_PAGE_SIZE = 50

// default history initial candles amount if not specified
const DEFAULT_INITIAL_HISTORY_CANDLES = 10
