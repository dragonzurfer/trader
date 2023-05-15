package atmcs

import (
	"errors"
	"fmt"
	"log"

	"github.com/dragonzurfer/trader/atmcs/trade"
	"github.com/dragonzurfer/trader/executor"
)

func (obj *ATMcs) ExitAccount() {}

func (obj *ATMcs) ExitPaper() {
	if !obj.Trade.InTrade {
		return
	}

	exitPositions, err := obj.MakeExitPositions()
	if err != nil {
		log.Println(err.Error())
		return
	}

	// Clear the current trade
	obj.Trade.InTrade = false
	obj.Trade.ExitPositions = exitPositions
	obj.Trade.TimeOfExit = obj.GetCurrentTime()
}

func (obj *ATMcs) MakeExitPositions() ([]trade.OptionPosition, error) {
	var exitPositions []trade.OptionPosition

	// Loop through the current entry positions and create corresponding exit positions
	for _, entryPosition := range obj.Trade.EntryPositions {
		position := entryPosition
		exitPosition, err := obj.MakePositionExit(position)
		if err != nil {
			return nil, errors.New("failed to makeExitPositions():" + err.Error())
		}
		exitPositions = append(exitPositions, exitPosition)
	}

	return exitPositions, nil
}

func (obj *ATMcs) MakePositionExit(entryPosition trade.OptionPosition) (trade.OptionPosition, error) {
	// Fetch the current market price for the option
	var depth []executor.MarketDepthLike
	var err error
	var exitPosition trade.OptionPosition
	if entryPosition.GetTradeType() == executor.Buy {
		depth, err = GetBids(obj.Broker, entryPosition)
	} else {
		depth, err = GetAsks(obj.Broker, entryPosition)
	}

	if err != nil {
		return exitPosition, errors.New("failed to MakeExitPosition():" + err.Error())
	}

	currentPrice := obj.GetAvgMarketDepth(depth)
	if currentPrice == 0 {
		return exitPosition, fmt.Errorf("could not get exit price for %v", entryPosition.GetOptionSymbol())

	}

	// Create a new option position with the current price and trade type reversed
	exitPosition = trade.OptionPosition{
		Option: trade.Option{
			Strike:           entryPosition.GetStrike(),
			Expiry:           entryPosition.GetExpiry(),
			Type:             entryPosition.GetOptionType(),
			Symbol:           entryPosition.GetOptionSymbol(),
			UnderlyingSymbol: entryPosition.GetUnderlyingSymbol(),
		},
		TradeType: reverseTradeType(entryPosition.GetTradeType()),
		Price:     currentPrice,
		Quantity:  entryPosition.GetQuantity(),
	}

	return exitPosition, nil
}

func reverseTradeType(tradeType executor.TradeType) executor.TradeType {
	if tradeType == executor.Buy {
		return executor.Sell
	} else if tradeType == executor.Sell {
		return executor.Buy
	} else {
		return executor.Nuetral
	}
}
