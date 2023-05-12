package executor

import (
	"time"
)

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

type TradeType string

const (
	Sell    TradeType = "Sell"
	Buy     TradeType = "Buy"
	Nuetral TradeType = "Neutral"
)

type TimeFrame string

const (
	Day      TimeFrame = "1Day"
	Minute   TimeFrame = "min"
	Minute5  TimeFrame = "5min"
	Minute15 TimeFrame = "15min"
	Minute30 TimeFrame = "30min"
	Minute45 TimeFrame = "45min"
	Hour     TimeFrame = "1hour"
	Hour4    TimeFrame = "4hour"
	Month    TimeFrame = "1month"
	Month3   TimeFrame = "3month"
	Month6   TimeFrame = "6month"
	Year     TimeFrame = "1year"
)

type OptionLike interface {
	GetExpiry() time.Time
	GetStrike() float64
	GetOptionType() OptionType
	GetOptionSymbol() string
	GetUnderlyingSymbol() string
}

type Expiry struct {
	ExpiryDate time.Time
}

type OptionPositionLike interface {
	OptionLike
	GetTradeType() TradeType
	GetPrice() float64
	GetQuantity() int64
}

type BrokerLike interface {
	SetCredentialsFilePath(string)
	GetLTP(string) (float64, error)
	GetMarketDepth(string) (BidAskLike, error)
	GetCandles(string, time.Time, time.Time, TimeFrame) ([]CandleLike, error)
	GetOptionExpiries(string) ([]Expiry, error)
	GetMarketDepthOption(float64, time.Time, OptionType) (BidAskLike, error)
	GetCandlesOption(float64, time.Time, OptionType, time.Time, time.Time) ([]CandleLike, error)
}
