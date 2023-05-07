package executor

import "time"

type MarketDepthLike interface {
	GetPrice() float64
	GetQuantity() int64
	GetNumOfOrders() int64
}

type BidAskLike interface {
	GetBids() []MarketDepthLike
	GetAsks() []MarketDepthLike
}

type CandleLike interface {
	GetOpen() float64
	GetHigh() float64
	GetLow() float64
	GetClose() float64
	GetVolume() float64
	GetOI() float64
}

type OptionType string

const (
	CallOption OptionType = "CE"
	PutOption  OptionType = "PE"
)

type OptionLike interface {
	GetExpiry() time.Time
	GetStrike() float64
	GetOptionType() OptionType
	GetOptionSymbol() string
	GetUnderlyingSymbol() string
}

type OptionChainLike interface {
	SortStrikeLowToHigh()
	GetOptions() []OptionLike
	GetExpiry() time.Time
	GetUnderlyingSymbol() string
}

type BrokerLike interface {
	SetCredentialsFilePath(string)
	GetLTP(string) (float64, error)
	GetMarketDepth(string) (BidAskLike, error)
	GetCandles(string) ([]CandleLike, error)
	GetOptionChain(string, string) OptionChainLike
}
