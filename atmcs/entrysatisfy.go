package atmcs

import (
	"errors"
	"log"
	"time"

	cpr "github.com/dragonzurfer/strategy/CPR"
	"github.com/dragonzurfer/trader/executor"
)

func (obj *ATMcs) IsEntrySatisfied() bool {
	minTargetPercent := obj.Settings.MinTargetPercent
	minSLPercent := obj.Settings.MinStopLossPercent
	currentDayCandles, previousDayCandle, err := obj.GetCandles()
	if err != nil {
		log.Println("error in IsEntrySatisfied():", err.Error())
		return false
	}

	obj.SignalCPR = cpr.GetCPRSignal(minSLPercent, minTargetPercent, previousDayCandle, currentDayCandles)
	obj.EntrySatisfied = false
	obj.Trade.TradeType = executor.Nuetral
	switch obj.SignalCPR.Signal {
	case cpr.Buy:
		obj.EntrySatisfied = true
		obj.Trade.TradeType = executor.Buy
	case cpr.Sell:
		obj.EntrySatisfied = true
		obj.Trade.TradeType = executor.Sell
	}
	return obj.EntrySatisfied
}

func (obj *ATMcs) GetCandles() (cpr.CPRCandles, cpr.CPRCandles, error) {
	currentTime := obj.GetCurrentTime()
	previousDate := obj.GetPreviousNonWeekendNonHolidayDate(currentTime)
	currentDay5minCandles, err := obj.GetCurrentDayCandleData5minFyers(currentTime)
	if err != nil {
		return nil, nil, errors.New("error in GetCandles() current day 5min candle data: " + err.Error())
	}
	previousDayCandles, err := obj.GetPreviousDayCandleDataFyers(previousDate)
	if err != nil {
		return nil, nil, errors.New("error in GetCandles() getting previous day candle data: " + err.Error())
	}
	if currentDay5minCandles.GetCandlesLength() < 1 {
		return nil, nil, errors.New("zero candles returned on 5minute data API call to broker")
	}
	if previousDayCandles.GetCandlesLength() < 1 {
		return nil, nil, errors.New("zero candles returned on 1Day data API call to broker")
	}
	return currentDay5minCandles, previousDayCandles, nil
}

func (obj *ATMcs) GetPreviousDayCandleDataFyers(previousDate time.Time) (cpr.CPRCandles, error) {
	from := time.Date(previousDate.Year(), previousDate.Month(), previousDate.Day(), 9, 15, 0, 0, obj.ISTLocation)
	to := time.Date(previousDate.Year(), previousDate.Month(), previousDate.Day(), 15, 30, 0, 0, obj.ISTLocation)
	candles, err := obj.Broker.GetCandles(obj.Symbol, from, to, executor.Day)
	if err != nil {
		return nil, err
	}
	cprCandles := NewCandleAdapter(candles)
	return cprCandles, err
}

func (obj *ATMcs) GetCurrentDayCandleData5minFyers(currentTime time.Time) (cpr.CPRCandles, error) {
	from := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 9, 15, 0, 0, obj.ISTLocation)
	to := currentTime
	candles, err := obj.Broker.GetCandles(obj.Symbol, from, to, executor.Minute5)
	if err != nil {
		return nil, err
	}
	cprCandles := NewCandleAdapter(candles)
	return cprCandles, nil
}

type Holidays struct {
	HolidayDates []string `json:"holiday_dates"`
}

func isDateHoliday(date string, holidays *Holidays) bool {
	for _, holiday := range holidays.HolidayDates {
		if holiday == date {
			return true
		}
	}
	return false
}

func WasMarketOpen(date time.Time, holidays *Holidays) bool {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Println("Error loading location:", err)
		return false
	}

	// Check if the date is a weekend (Saturday or Sunday)
	if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		return false
	}

	dateString := date.In(loc).Format("2006-01-02")

	// Check if the date is a holiday
	if isDateHoliday(dateString, holidays) {
		return false
	}

	return true
}

func (obj *ATMcs) GetPreviousNonWeekendNonHolidayDate(currentDate time.Time) time.Time {
	previousDate := currentDate.AddDate(0, 0, -1)
	for !WasMarketOpen(previousDate, &obj.Holidays) {
		previousDate = previousDate.AddDate(0, 0, -1)
	}
	return previousDate
}
