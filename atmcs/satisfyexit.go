package atmcs

import (
	"github.com/dragonzurfer/trader/executor"
)

// func (obj *ATMcs) SubscribeCheckSL() {

// 	onMessageFunc := func(notification api.Notification) {
// 		fmt.Println(notification.Type, notification.SymbolData)
// 		obj.SetMinTrail(float64(notification.SymbolData.Ltp))
// 		isHitTickSL := obj.IsHitTickSL(float64(notification.SymbolData.Ltp))
// 		if isHitTickSL {
// 			// send to cahnnel
// 			obj.StopLossHitChan <- true
// 		}
// 	}

// 	onErrorFunc := func(err error) {
// 		log.Errorf("failed to watch | disconnected from watch. %v", err)
// 	}

// 	onCloseFunc := func() {
// 		log.Println("watch connection is closed")
// 	}

// 	cli := fyerswatch.New()
// 	WithOnMessageFunc(onMessageFunc).
// 		WithOnErrorFunc(onErrorFunc).
// 		WithOnCloseFunc(onCloseFunc)

// 	// this is a blocking call
// 	cli.Subscribe(api.SymbolDataTick, obj.Symbol)
// }

// will be called while creating broker object
func (obj *ATMcs) IsHitTickSL(tickPrice float64) bool {
	if !obj.InTrade() {
		return false
	}
	switch obj.Trade.TradeType {
	case executor.Buy:
		return tickPrice <= obj.Trade.StopLossPrice
	case executor.Sell:
		return tickPrice >= obj.Trade.StopLossPrice
	}
	return false
}

func (obj *ATMcs) SetMinTrail(tickPrice float64) {
	// Check if a trade is active
	if !obj.InTrade() {
		obj.Trade.IsMinTrailHit = false
	}

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
			obj.Trade.StopLossPrice = obj.Trade.EntryPrice
		}
	}
}
