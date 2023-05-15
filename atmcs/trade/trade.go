package trade

import (
	"time"

	"github.com/dragonzurfer/trader/executor"
)

type Trade struct {
	InTrade            bool
	EntryPositions     []executor.OptionPositionLike
	ExitPositions      []executor.OptionPositionLike
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

func (t *Trade) GetEntryPositions() []executor.OptionPositionLike {
	return t.EntryPositions
}

func (t *Trade) SetEntryPositions(positions []executor.OptionPositionLike) {
	t.EntryPositions = positions
}

func (t *Trade) GetExitPositions() []executor.OptionPositionLike {
	return t.ExitPositions
}

func (t *Trade) SetExitPositions(positions []executor.OptionPositionLike) {
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
