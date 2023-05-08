package trade

import (
	"time"

	"github.com/dragonzurfer/trader/executor"
)

type Trade struct {
	EntryPositions []executor.OptionLike
	ExitPositions  []executor.OptionLike
	TimeOfEntry    time.Time
	TimeOfExit     time.Time
	StopLossPrice  float64
	TargetPrice    float64
}

func (t *Trade) GetEntryPositions() []executor.OptionLike {
	return t.EntryPositions
}

func (t *Trade) SetEntryPositions(positions []executor.OptionLike) {
	t.EntryPositions = positions
}

func (t *Trade) GetExitPositions() []executor.OptionLike {
	return t.ExitPositions
}

func (t *Trade) SetExitPositions(positions []executor.OptionLike) {
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
