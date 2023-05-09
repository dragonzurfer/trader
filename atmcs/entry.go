package atmcs

import (
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/dragonzurfer/trader/atmcs/trade"
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

func (obj *ATMcs) IsEntrySatisfied() bool {
	return true
}

func (obj *ATMcs) PaperTrade(tradeType executor.TradeType) {
	obj.Trade = trade.Trade{
		InTrade:        true,
		EntryPositions: obj.makeEntryPositions(tradeType),
	}
}

func (obj *ATMcs) makeEntryPositions(tradeType executor.TradeType) []executor.OptionLike {
	ltp, err := obj.Broker.GetLTP(obj.Symbol)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	atmStrike := GetStrikeATM(ltp, obj.StrikeDiff)
	strike := GetNearest100ITMStrike(atmStrike, tradeType)

	expiries, err := obj.Broker.GetOptionExpiries(obj.Symbol)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	sellExpiry, err := GetExpiry(obj.GetCurrentTime(), obj.MinDaysToExpiry, strike, expiries)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	buyExpiry, err := GetMonthlyExpiry(obj.GetCurrentTime(), expiries)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	symbol := obj.Symbol

	quantity := obj.Quantity
	var sellPosition, buyPosition OptionPosition
	var entryPositions []executor.OptionLike
	if tradeType == executor.Buy {
		sellPosition = MakeEntryPosition(symbol, strike, sellExpiry, executor.PutOption, executor.Sell, quantity)
		buyPosition = MakeEntryPosition(symbol, strike, buyExpiry, executor.PutOption, executor.Buy, quantity/2)
	} else {
		sellPosition = MakeEntryPosition(symbol, strike, sellExpiry, executor.CallOption, executor.Sell, quantity)
		buyPosition = MakeEntryPosition(symbol, strike, buyExpiry, executor.CallOption, executor.Buy, quantity/2)
	}
	bids, err := GetBids(obj.Broker, sellPosition)
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	asks, err := GetAsks(obj.Broker, buyPosition)
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	sellPosition.Price = GetAvgMarketDepth(bids)
	buyPosition.Price = GetAvgMarketDepth(asks)
	entryPositions = append(entryPositions, sellPosition, buyPosition)
	return entryPositions
}

func MakeEntryPosition(symbol string, strike float64, expiry executor.Expiry, optionType executor.OptionType, tradeType executor.TradeType, quantity int64) OptionPosition {
	option := Option{
		Strike:           strike,
		Expiry:           expiry.ExpiryDate,
		Type:             optionType,
		Symbol:           "NIFTY", //change to broker symbol
		UnderlyingSymbol: symbol,
	}
	optionPosition := OptionPosition{
		Option:    option,
		TradeType: tradeType,
		Quantity:  quantity,
	}
	return optionPosition
}

func GetAvgMarketDepth(depth []executor.MarketDepthLike) float64 {
	if len(depth) == 0 {
		return 0
	}

	total := 0.0
	totalVolumeOrders := 0.0
	for _, marketDepth := range depth {
		volumeOrders := float64(marketDepth.GetQuantity() * marketDepth.GetNumOfOrders())
		total += marketDepth.GetPrice() * volumeOrders
		totalVolumeOrders += volumeOrders
	}

	if totalVolumeOrders == 0 {
		return 0
	}

	avgPrice := total / totalVolumeOrders
	roundedPrice := roundToNearest(avgPrice, 0.05)

	return roundedPrice
}

func roundToNearest(val, roundTo float64) float64 {
	rounded := roundTo * float64(int(val/roundTo+0.5))
	return rounded
}

func GetBids(broker executor.BrokerLike, pos OptionPosition) ([]executor.MarketDepthLike, error) {
	bid_aks, err := broker.GetMarketDepthOption(pos.Strike, pos.Expiry, pos.Type)
	return bid_aks.GetBids(), err
}

func GetAsks(broker executor.BrokerLike, pos OptionPosition) ([]executor.MarketDepthLike, error) {
	bid_aks, err := broker.GetMarketDepthOption(pos.Strike, pos.Expiry, pos.Type)
	return bid_aks.GetAsks(), err
}

func GetExpiry(currentTime time.Time, minDaysToExpiry int64, strike float64, expiries []executor.Expiry) (executor.Expiry, error) {
	err := errors.New(fmt.Sprintf("Could not find an expiry that has %v days to expiry", minDaysToExpiry))
	minDiff := int64(math.MaxInt64)
	var earliestExpiry executor.Expiry
	for _, expiry := range expiries {
		daysDiff := int64(expiry.ExpiryDate.Sub(currentTime).Hours() / 24)
		if daysDiff >= minDaysToExpiry && daysDiff < minDiff {
			minDiff = daysDiff
			earliestExpiry = expiry
			err = nil
		}
	}

	return earliestExpiry, err
}

func GetMonthlyExpiry(currentTime time.Time, expiries []executor.Expiry) (executor.Expiry, error) {
	currentYear, currentMonth := currentTime.Year(), currentTime.Month()

	var lastExpiryOfCurrentMonth executor.Expiry
	var lastExpiryOfNextMonth executor.Expiry
	var foundCurrentMonth, foundNextMonth bool

	nextMonth := currentMonth + 1
	nextYear := currentYear

	if nextMonth > 12 {
		nextMonth = 1
		nextYear++
	}

	for _, expiry := range expiries {
		expiryYear, expiryMonth := expiry.ExpiryDate.Year(), expiry.ExpiryDate.Month()

		if expiryYear == currentYear && expiryMonth == currentMonth {
			lastExpiryOfCurrentMonth = expiry
			foundCurrentMonth = true
		} else if expiryYear == nextYear && expiryMonth == nextMonth {
			lastExpiryOfNextMonth = expiry
			foundNextMonth = true
		} else if expiryYear > nextYear || (expiryYear == nextYear && expiryMonth > nextMonth) {
			break
		}
	}

	if foundCurrentMonth {
		return lastExpiryOfCurrentMonth, nil
	} else if foundNextMonth {
		return lastExpiryOfNextMonth, nil
	} else {
		return executor.Expiry{}, errors.New("no suitable monthly expiry found")
	}
}
func GetStrikeATM(ltp float64, strikeDiff float64) float64 {
	return roundToNearestMultiple(ltp, strikeDiff)
}
func roundToNearestMultiple(value, multiple float64) float64 {
	return multiple * math.Round(value/multiple)
}

func GetNearest100ITMStrike(ltp float64, tradeType executor.TradeType) float64 {
	closest100Multiple := roundToNearest100Multiple(ltp)

	switch tradeType {
	case executor.Buy:
		if closest100Multiple < ltp {
			return closest100Multiple + 100
		}
		return closest100Multiple
	case executor.Sell:
		if closest100Multiple > ltp {
			return closest100Multiple - 100
		}
		return closest100Multiple
	default:
		panic("Invalid trade type")
	}
}

func roundToNearest100Multiple(value float64) float64 {
	return 100 * math.Round(value/100)
}