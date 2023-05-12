package atmcs_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	fyers "github.com/dragonzurfer/fyersgo"
	"github.com/dragonzurfer/fyersgo/api"
	"github.com/dragonzurfer/trader/atmcs"
	"github.com/dragonzurfer/trader/executor"
)

type BrokerCandleLikeAdapter struct {
	candle api.Candle
}

func (cla *BrokerCandleLikeAdapter) GetOpen() float64 {
	return float64(cla.candle.OpenValue)
}

func (cla *BrokerCandleLikeAdapter) GetHigh() float64 {
	return float64(cla.candle.HighestValue)
}

func (cla *BrokerCandleLikeAdapter) GetLow() float64 {
	return float64(cla.candle.LowestValue)
}

func (cla *BrokerCandleLikeAdapter) GetClose() float64 {
	return float64(cla.candle.CloseValue)
}

func (cla *BrokerCandleLikeAdapter) GetVolume() float64 {
	return float64(cla.candle.Volume)
}

func (cla *BrokerCandleLikeAdapter) GetOI() float64 {
	// The original API does not provide OI (Open Interest), so we return a default value.
	return 0
}

func NewBrokerCandleLikeAdapter(candles []api.Candle) []executor.CandleLike {
	candleLike := make([]executor.CandleLike, len(candles))
	for i, c := range candles {
		candleLike[i] = &BrokerCandleLikeAdapter{candle: c}
	}
	return candleLike
}

type RealBroker struct {
	LTP         float64
	Expiries    []executor.Expiry
	BidAsks     map[time.Time]BidAsk
	ApiKey      string
	AccessToken string
}

func (b *RealBroker) SetCredentialsFilePath(string) {}

func (b *RealBroker) GetLTP(string) (float64, error) {
	return b.LTP, nil
}

func (b *RealBroker) GetMarketDepth(string) (executor.BidAskLike, error) {
	return b.BidAsks[time.Now()], nil
}

func (b *RealBroker) GetCandles(symbol string, startTime time.Time, endTime time.Time) ([]executor.CandleLike, error) {
	cli := fyers.New(b.ApiKey, b.AccessToken)
	data, err := cli.GetHistoricalData(symbol, api.Minute5, startTime, endTime)
	if err != nil {
		return nil, errors.New("failed to get candles from fyers:" + err.Error())
	}
	candles := NewBrokerCandleLikeAdapter(data.Candles)
	for _, candle := range candles {
		fmt.Println(candle.GetHigh(), candle.GetLow(), candle.GetClose())
	}
	return candles, nil
}

func (b *RealBroker) GetOptionExpiries(string) ([]executor.Expiry, error) {
	return b.Expiries, nil
}

func (b *RealBroker) GetMarketDepthOption(strike float64, optExpiry time.Time, optType executor.OptionType) (executor.BidAskLike, error) {
	for expiry, bidask := range b.BidAsks {
		if expiry == optExpiry {
			return bidask, nil
		}
	}
	return nil, errors.New("API Could not find expiry date")
}

func (b *RealBroker) GetCandlesOption(float64, time.Time, executor.OptionType, time.Time, time.Time) ([]executor.CandleLike, error) {
	return nil, nil
}

func NewRealBroker(t *testing.T) RealBroker {
	type Creds struct {
		AccessToken string `json:"access_token"`
		ApiKey      string `json:"api_key"`
	}
	var cred Creds
	currentFilePath, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting current file path: %v", err)
	}
	credentialsPath := filepath.Join(currentFilePath, "credentials.json")
	data, err := ioutil.ReadFile(credentialsPath)
	if err != nil {
		t.Fatalf("Error reading test case file: %v", err)
	}
	json.Unmarshal(data, &cred)

	broker := RealBroker{
		ApiKey:      cred.ApiKey,
		AccessToken: cred.AccessToken,
	}
	return broker
}

func TestIsEntrySatisfied(t *testing.T) {
	LoadTimeLocation()
	log.SetFlags(log.Lshortfile)
	historicalTime := []string{
		"2023-05-02T10:15:00+05:30",
		"2023-05-11T13:34:59+05:30",
		"2023-05-12T09:45:00+05:30",
	}
	settings := []string{
		"testcase1settings.json",
		"testcase1settings.json",
		"testcase1settings.json",
	}

	for i, timeString := range historicalTime {
		fmt.Println("Running for time: ", timeString)
		timeObj, err := time.Parse("2006-01-02T15:04:05-07:00", timeString)
		if err != nil {
			t.Fatal("Error parsing time:", err)
			return
		}

		settingsFileName := settings[i]
		currentFilePath, err := os.Getwd()
		if err != nil {
			t.Fatalf("Error getting current file path: %v", err)
		}
		settingsFilePath := filepath.Join(currentFilePath, "testcases", "entrysatisfy", settingsFileName)
		var actualObj executor.ExecutorLike
		actualObj = atmcs.New(settingsFilePath, currentFilePath, func() time.Time { return timeObj })
		if actualObj == nil {
			t.Fatalf("ATMcs object init fail")
		}
		broker := NewRealBroker(t)
		actualObj.SetBroker(&broker)
		fmt.Println(actualObj.GetTradeType())
		if actualObj.IsEntrySatisfied() != false {
			t.Fatalf("entry is satisfied:%+v", actualObj.GetTradeType())
		}
	}
	
}
