package atmcs

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
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

func (obj *ATMcs) PaperTrade(tradeType executor.TradeType) {
	obj.Trade = trade.Trade{
		InTrade:        true,
		EntryPositions: obj.makeEntryPositions(tradeType),
		TimeOfEntry:    obj.GetCurrentTime(),
	}
}

func (obj *ATMcs) makeEntryPositions(tradeType executor.TradeType) []executor.OptionPositionLike {

	ltp, err := obj.Broker.GetLTP(obj.Symbol)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	strike := GetNearest100ITMStrike(ltp, tradeType)

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

	buyExpiry, err := GetMonthlyExpiryCalendarSpread(obj.GetCurrentTime(), sellExpiry.ExpiryDate, expiries)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	symbol := obj.Symbol

	quantity := obj.Quantity
	var sellPosition, buyPosition OptionPosition
	var entryPositions []executor.OptionPositionLike
	if tradeType == executor.Buy {
		sellPosition = obj.MakeEntryPosition(symbol, strike, sellExpiry, executor.PutOption, executor.Sell, quantity)
		buyPosition = obj.MakeEntryPosition(symbol, strike, buyExpiry, executor.PutOption, executor.Buy, quantity/2)
	} else {
		sellPosition = obj.MakeEntryPosition(symbol, strike, sellExpiry, executor.CallOption, executor.Sell, quantity)
		buyPosition = obj.MakeEntryPosition(symbol, strike, buyExpiry, executor.CallOption, executor.Buy, quantity/2)
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
	sellPosition.Price = obj.GetAvgMarketDepth(bids)
	buyPosition.Price = obj.GetAvgMarketDepth(asks)
	entryPositions = append(entryPositions, sellPosition, buyPosition)
	return entryPositions
}

func (obj *ATMcs) MakeEntryPosition(symbol string, strike float64, expiry executor.Expiry, optionType executor.OptionType, tradeType executor.TradeType, quantity int64) OptionPosition {
	option := Option{
		Strike:           strike,
		Expiry:           expiry.ExpiryDate,
		Type:             optionType,
		Symbol:           obj.Symbol, //change to broker symbol
		UnderlyingSymbol: symbol,
	}
	optionPosition := OptionPosition{
		Option:    option,
		TradeType: tradeType,
		Quantity:  quantity,
	}
	return optionPosition
}

func (obj *ATMcs) GetAvgMarketDepth(depth []executor.MarketDepthLike) float64 {
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
	roundedPrice := roundToNearest(avgPrice, obj.Settings.TickSize)
	if roundedPrice < avgPrice {
		roundedPrice += obj.Settings.TickSize
	}
	decimalPlaces := countDecimalPlaces(obj.Settings.TickSize)
	roundedPrice = truncateDecimal(roundedPrice, decimalPlaces)
	return roundedPrice
}

func GetBids(broker executor.BrokerLike, pos OptionPosition) ([]executor.MarketDepthLike, error) {
	bid_aks, err := broker.GetMarketDepthOption(pos.Strike, pos.Expiry, pos.Type)
	if err != nil {
		return nil, err
	}
	return bid_aks.GetBids(), err
}

func GetAsks(broker executor.BrokerLike, pos OptionPosition) ([]executor.MarketDepthLike, error) {
	bid_aks, err := broker.GetMarketDepthOption(pos.Strike, pos.Expiry, pos.Type)
	if err != nil {
		return nil, err
	}
	return bid_aks.GetAsks(), err
}

func GetExpiry(currentTime time.Time, minDaysToExpiry int64, strike float64, expiries []executor.Expiry) (executor.Expiry, error) {
	err := fmt.Errorf("Could not find an expiry that has %v days to expiry", minDaysToExpiry)
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

func GetMonthlyExpiryCalendarSpread(currentTime time.Time, sellExpiry time.Time, expiries []executor.Expiry) (executor.Expiry, error) {
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
		if expiry.ExpiryDate.After(sellExpiry) {

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
	}

	if foundCurrentMonth {
		return lastExpiryOfCurrentMonth, nil
	} else if foundNextMonth {
		return lastExpiryOfNextMonth, nil
	} else {
		return executor.Expiry{}, errors.New("no suitable monthly expiry found")
	}
}

func GetNearest100ITMStrike(ltp float64, tradeType executor.TradeType) float64 {
	closest100Multiple := roundToNearest(ltp, 100)

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
	}
	return closest100Multiple
}

func roundToNearest(val, roundTo float64) float64 {
	if roundTo == 0 {
		return val
	}
	rounded := roundTo * float64(int(val/roundTo))
	return rounded
}

func countDecimalPlaces(val float64) int {
	str := strconv.FormatFloat(val, 'f', -1, 64)
	parts := strings.Split(str, ".")
	if len(parts) == 2 {
		return len(parts[1])
	}
	return 0
}

func truncateDecimal(val float64, decimalPlaces int) float64 {
	shift := math.Pow(10, float64(decimalPlaces))
	truncated := math.Floor(val*shift) / shift

	return truncated
}
