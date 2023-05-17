package atmcs

import (
	"fmt"
	"strings"
)

func (obj *ATMcs) GetEntryMessage() string {
	obj.SetEntryMessage()
	return obj.EntryMessage
}

func (obj *ATMcs) GetExitMessage() string {
	obj.SetExitMessage()
	return obj.ExitMessage
}

func (obj *ATMcs) SetEntryMessage() {
	entryPositions := obj.Trade.EntryPositions
	trade := obj.Trade

	var messages []string
	underLyingEntryMsg := fmt.Sprintf(
		"%v %v\nEntry: %.2f\nSL: %.2f\nTarget: %.2f\n",
		obj.Symbol,
		obj.Trade.TradeType,
		obj.Trade.EntryPrice,
		obj.Trade.StopLossPrice,
		obj.Trade.TargetPrice,
	)
	messages = append(messages, underLyingEntryMsg)
	for _, position := range entryPositions {
		message := fmt.Sprintf(
			"%v %v at price %.2f\nexpiry: %v\nquantity: %d\n",
			position.GetStrike(),
			position.GetOptionType(),
			position.GetPrice(),
			position.GetExpiry().Format("2006-01-02"),
			position.GetQuantity(),
		)
		messages = append(messages, message)
	}
	depthQuantMessage := fmt.Sprintf("depth Quant sell enter:%0.2f depth Quant buy enter:%0.2f", obj.Trade.DepthQuantityEntrySell, obj.Trade.DepthQuantityEntryBuy)
	messages = append(messages, depthQuantMessage)
	entryTimeMsg := fmt.Sprintf("Entry Time:%v", trade.TimeOfEntry.Format("2006-01-02 15:04:05"))
	messages = append(messages, entryTimeMsg)

	obj.EntryMessage = strings.Join(messages, "\n")
}

func (obj *ATMcs) SetExitMessage() {
	exitPositions := obj.Trade.ExitPositions
	trade := obj.Trade

	var messages []string
	for _, position := range exitPositions {
		message := fmt.Sprintf(
			"Exit %v %v at price %.2f\nexpiry: %v\nquantity: %d\n",
			position.GetStrike(),
			position.GetOptionType(),
			position.GetPrice(),
			position.GetExpiry().Format("2006-01-02"),
			position.GetQuantity(),
		)
		messages = append(messages, message)
	}
	depthQuantMessage := fmt.Sprintf("depth Quant sell enter:%0.2f depth Quant buy enter:%0.2f", obj.Trade.DepthQuantityEntrySell, obj.Trade.DepthQuantityEntryBuy)
	messages = append(messages, depthQuantMessage)
	exitTimeMsg := fmt.Sprintf("Exit Time:%v", trade.TimeOfExit.Format("2006-01-02 15:04:05"))
	messages = append(messages, exitTimeMsg)

	obj.ExitMessage = strings.Join(messages, "\n")
}

func (obj *ATMcs) GetCSVTradeEntry() []string {
	entry := obj.Trade.EntryPositions
	entryStrings := make([]string, len(entry))

	for i, position := range entry {
		entryStrings[i] = fmt.Sprintf(
			"Type:Enter, Strike: %.2f, OptionType: %s, Price: %.2f, TradeType: %s, Quantity: %d, Expiry: %s, TimeOfEntry: %s, TimeOfExit: %s,Available SellQty: %.2f, Available BuyQty: %.2f",
			position.GetStrike(),
			position.GetOptionType(),
			position.GetPrice(),
			position.GetTradeType(),
			position.GetQuantity(),
			position.GetExpiry().Format("2006-01-02"),
			obj.Trade.TimeOfEntry.Format("2006-01-02 15:04:05"),
			obj.Trade.TimeOfExit.Format("2006-01-02 15:04:05"),
			obj.Trade.DepthQuantityEntrySell,
			obj.Trade.DepthQuantityEntryBuy,
		)
	}

	return entryStrings
}

func (obj *ATMcs) GetCSVTradeExit() []string {
	exit := obj.Trade.ExitPositions
	exitStrings := make([]string, len(exit))

	for i, position := range exit {
		exitStrings[i] = fmt.Sprintf(
			"Type:Exit, Strike: %.2f, OptionType: %s, Price: %.2f, TradeType: %s, Quantity: %d, Expiry: %s, TimeOfEntry: %s, TimeOfExit: %s,Available SellQty: %.2f, Available BuyQty: %.2f",
			position.GetStrike(),
			position.GetOptionType(),
			position.GetPrice(),
			position.GetTradeType(),
			position.GetQuantity(),
			position.GetExpiry().Format("2006-01-02"),
			obj.Trade.TimeOfEntry.Format("2006-01-02 15:04:05"),
			obj.Trade.TimeOfExit.Format("2006-01-02 15:04:05"),
			obj.Trade.DepthQuantityExitSell,
			obj.Trade.DepthQuantityExitBuy,
		)
	}

	return exitStrings
}
