package trade

import (
	"time"

	"github.com/dragonzurfer/trader/executor"
)

type Option struct {
	Expiry           time.Time           `json:"expiry"`
	Strike           float64             `json:"strike"`
	Type             executor.OptionType `json:"type"`
	Symbol           string              `json:"symbol"`
	UnderlyingSymbol string              `json:"underlying_symbol"`
}

type OptionPosition struct {
	Option
	Price     float64
	TradeType executor.TradeType
	Quantity  int64
}

func (op OptionPosition) GetTradeType() executor.TradeType { return op.TradeType }
func (op OptionPosition) GetPrice() float64                { return op.Price }
func (op OptionPosition) GetQuantity() int64               { return op.Quantity }

func (o Option) GetExpiry() time.Time               { return o.Expiry }
func (o Option) GetStrike() float64                 { return o.Strike }
func (o Option) GetOptionType() executor.OptionType { return o.Type }
func (o Option) GetOptionSymbol() string            { return o.Symbol }
func (o Option) GetUnderlyingSymbol() string        { return o.UnderlyingSymbol }

type Trade struct {
	InTrade            bool
	EntryPositions     []OptionPosition
	ExitPositions      []OptionPosition
	TimeOfEntry        time.Time
	TimeOfExit         time.Time
	TrailStopLossPrice float64
	StopLossPrice      float64
	TargetPrice        float64
	EntryPrice         float64
	TradeType          executor.TradeType
	IsMinTrailHit      bool
	IsStopLossHit      bool
}

func (t *Trade) GetEntryPositions() []OptionPosition {
	return t.EntryPositions
}

func (t *Trade) SetEntryPositions(positions []OptionPosition) {
	t.EntryPositions = positions
}

func (t *Trade) GetExitPositions() []OptionPosition {
	return t.ExitPositions
}

func (t *Trade) SetExitPositions(positions []OptionPosition) {
	t.ExitPositions = positions
}

func (t *Trade) GetTimeOfEntry() time.Time {
	return t.TimeOfEntry
}

func (t *Trade) SetTimeOfEntry(timeOfEntry time.Time) {
	t.TimeOfEntry = timeOfEntry
}

func (t *Trade) GetTimeOfExit() time.Time {
	return t.TimeOfExit
}

func (t *Trade) SetTimeOfExit(timeOfExit time.Time) {
	t.TimeOfExit = timeOfExit
}

func (t *Trade) GetStopLossPrice() float64 {
	return t.StopLossPrice
}

func (t *Trade) SetStopLossPrice(stopLossPrice float64) {
	t.StopLossPrice = stopLossPrice
}

func (t *Trade) GetTargetPrice() float64 {
	return t.TargetPrice
}

func (t *Trade) SetTargetPrice(targetPrice float64) {
	t.TargetPrice = targetPrice
}
