package atmcs

import (
	"fmt"

	"github.com/dragonzurfer/trader/executor"
)

func (obj *ATMcs) UpdateSL() {

}

func (obj *ATMcs) ExitOnTick(tickPrice float64) {
	if obj.InTrade() {
		obj.SetMinTrail(tickPrice)
		if obj.IsHitTickSL(tickPrice) {
			obj.ExitPaper()
			go func() {
				obj.StopLossHitChan <- true
			}()
			obj.Trade.IsStopLossHit = true
			obj.ExitSatisfied = true
			obj.Trade.TimeOfExit = obj.GetCurrentTime()
			return
		}
		if obj.IsHitTickTarget(tickPrice) {
			obj.ExitPaper()
			go func() {
				obj.TargetHitChan <- true
			}()
			obj.ExitSatisfied = true
			obj.Trade.TimeOfExit = obj.GetCurrentTime()
			return
		}
		if obj.IsUpdateMinTrail(tickPrice) {
			go func() {
				obj.TrailChan <- true
			}()
			return
		}
	}
}
func (obj *ATMcs) IsUpdateMinTrail(tickPrice float64) bool {
	if !obj.Trade.IsMinTrailHit {
		return false
	}
	if obj.Trade.StopLossPrice == obj.Trade.EntryPrice {
		return false
	}
	switch obj.Trade.TradeType {
	case executor.Buy:
		if tickPrice >= obj.Trade.TrailStopLossPrice {
			obj.Trade.StopLossPrice = obj.Trade.EntryPrice
			return true
		} else {
			return false
		}
	case executor.Sell:
		if tickPrice <= obj.Trade.TrailStopLossPrice {
			obj.Trade.StopLossPrice = obj.Trade.EntryPrice
			return true
		} else {
			return false
		}
	}
	return false
}

// will be called while creating broker object
func (obj *ATMcs) IsHitTickSL(tickPrice float64) bool {

	switch obj.Trade.TradeType {
	case executor.Buy:
		return tickPrice <= obj.Trade.StopLossPrice
	case executor.Sell:
		return tickPrice >= obj.Trade.StopLossPrice
	}
	return false
}

func (obj *ATMcs) IsHitTickTarget(tickPrice float64) bool {
	switch obj.Trade.TradeType {
	case executor.Buy:
		return tickPrice >= obj.Trade.TargetPrice
	case executor.Sell:
		return tickPrice <= obj.Trade.TargetPrice
	}
	fmt.Println(tickPrice, obj.Trade.TargetPrice)
	return false
}

func (obj *ATMcs) SetMinTrail(tickPrice float64) {
	// Calculate the absolute price change as a percentage
	priceChangePercent := 0.0
	if obj.Trade.TradeType == executor.Buy {
		// For a buy trade, we want the current price to be higher
		if tickPrice > obj.Trade.EntryPrice {
			priceChangePercent = (tickPrice - obj.Trade.EntryPrice) / obj.Trade.EntryPrice * 100
		}
	} else if obj.Trade.TradeType == executor.Sell {
		// For a sell trade, we want the current price to be lower
		if tickPrice < obj.Trade.EntryPrice {
			priceChangePercent = (obj.Trade.EntryPrice - tickPrice) / obj.Trade.EntryPrice * 100
		}
	}

	if priceChangePercent >= obj.Settings.MinTrailPercent {
		if !obj.Trade.IsMinTrailHit {
			obj.Trade.IsMinTrailHit = true
			if obj.Trade.TradeType == executor.Buy {
				obj.Trade.TrailStopLossPrice = obj.Trade.EntryPrice + (obj.Trade.EntryPrice - obj.Trade.StopLossPrice)
			}
			if obj.Trade.TradeType == executor.Sell {
				obj.Trade.TrailStopLossPrice = obj.Trade.EntryPrice - (obj.Trade.StopLossPrice - obj.Trade.EntryPrice)
			}
		}
	}
}
