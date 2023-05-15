package atmcs_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	fyers "github.com/dragonzurfer/fyersgo"
	"github.com/dragonzurfer/fyersgo/api"
	cpr "github.com/dragonzurfer/strategy/CPR"
	"github.com/dragonzurfer/trader/atmcs"
	"github.com/dragonzurfer/trader/executor"
)

var timeObj time.Time

type SceneTrade struct {
	InTrade            bool
	EntryPositions     []OptionPosition
	ExitPositions      []OptionPosition
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

type SceneATMcs struct {
	ISTLocation    *time.Location `json:"-"`
	SignalCPR      cpr.Signal
	EntrySatisfied bool
	ExitSatisfied  bool
	Broker         executor.BrokerLike
	EntryMessage   string
	ExitMessage    string
	SleepDuration  time.Duration
	Trade          SceneTrade
	Settings
	SettingsFilesPath string
	GetCurrentTime    func() time.Time `json:"-"`
	Holidays
	StopLossHitChan chan bool `json:"-"`
}

type TickPrice struct {
	Time  time.Time `json:"time"`
	Price float64   `json:"price"`
}

type TimeState struct {
	Time       time.Time  `json:"time"`
	ATMcsState SceneATMcs `json:"atmcs_state"`
}

type TestScenario struct {
	CurrentTime  time.Time            `json:"current_time"`
	TickPrices   []TickPrice          `json:"tick_prices"`
	TimeStates   []TimeState          `json:"time_states"`
	Expiries     []executor.Expiry    `json:"option_chains"`
	OptionDepths map[time.Time]BidAsk `json:"option_depths"`
}

type TestScenarioBroker struct {
	LTP         float64
	TimeChan    chan time.Time
	Expiries    []executor.Expiry
	BidAsks     map[time.Time]BidAsk
	ApiKey      string
	AccessToken string
}

func (b *TestScenarioBroker) SetCredentialsFilePath(string) {}

func (b *TestScenarioBroker) GetLTP(symbol string) (float64, error) {
	startTime := time.Date(timeObj.Year(), timeObj.Month(), timeObj.Day(), 9, 15, 0, 0, ISTLocation)
	candle, err := b.GetCandles(symbol, startTime, timeObj, executor.Minute5)
	if err != nil {
		return 0, fmt.Errorf("error getting ltp from broker GetLTP(string):%w", err)
	}
	if len(candle) < 1 {
		return 0, errors.New("empty candle data GetLTP(string)")
	}
	return candle[len(candle)-1].GetClose(), nil
}

func (b *TestScenarioBroker) GetMarketDepth(string) (executor.BidAskLike, error) {
	return b.BidAsks[time.Now()], nil
}

func (b *TestScenarioBroker) GetCandles(symbol string, startTime time.Time, endTime time.Time, tf executor.TimeFrame) ([]executor.CandleLike, error) {
	cli := fyers.New(b.ApiKey, b.AccessToken)
	var resolution api.Resolution
	switch tf {
	case executor.Minute5:
		resolution = api.Minute5
	case executor.Day:
		resolution = api.Day
	}
	data, err := cli.GetHistoricalData(symbol, resolution, startTime, endTime)
	if err != nil {
		return nil, errors.New("failed to get candles from fyers:" + err.Error())
	}

	index := 0
	if resolution == api.Minute5 {
		for i, c := range data.Candles {
			// fmt.Println(c.Timestamp, endTime)
			if c.Timestamp.After(endTime) {
				break
			}
			index = i
		}
	} else {
		index = len(data.Candles)
	}
	fmt.Println(startTime, ":", endTime)
	candles := NewBrokerCandleLikeAdapter(data.Candles[0:index])
	if len(candles) < 1 {
		return nil, errors.New("Emtpy candles BrokerLike.GetCandles")
	}
	return candles, nil
}

func (b *TestScenarioBroker) GetOptionExpiries(string) ([]executor.Expiry, error) {
	return b.Expiries, nil
}

func (b *TestScenarioBroker) GetMarketDepthOption(strike float64, optExpiry time.Time, optType executor.OptionType) (executor.BidAskLike, error) {
	for expiry, bidask := range b.BidAsks {
		if expiry == optExpiry {
			return bidask, nil
		}
	}
	return nil, errors.New("API Could not find expiry date")
}

func (b *TestScenarioBroker) GetCandlesOption(float64, time.Time, executor.OptionType, time.Time, time.Time) ([]executor.CandleLike, error) {
	return nil, nil
}

func NewScenarioBroker(t *testing.T) TestScenarioBroker {
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

	broker := TestScenarioBroker{
		ApiKey:      cred.ApiKey,
		AccessToken: cred.AccessToken,
	}
	return broker
}

func (e *SceneATMcs) StatesEqual(a *atmcs.ATMcs) bool {
	// fmt.Printf("expect:%v actual:%v\n", 1, 2)
	if math.Abs(e.SignalCPR.EntryPrice-a.SignalCPR.EntryPrice) >= 1 {
		fmt.Printf("EntryPrice Expect:%v Actual:%v\n", e.SignalCPR.EntryPrice, a.SignalCPR.EntryPrice)
		return false
	}
	if math.Abs(e.SignalCPR.StopLossPrice-a.SignalCPR.StopLossPrice) >= 1 {
		fmt.Printf("StopLossPrice Expect:%v Actual:%v\n", e.SignalCPR.StopLossPrice, a.SignalCPR.StopLossPrice)
		return false
	}
	if math.Abs(e.SignalCPR.TargetPrice-a.SignalCPR.TargetPrice) >= 1 {
		fmt.Printf("TargetPrice Expect:%v Actual:%v\n", e.SignalCPR.TargetPrice, a.SignalCPR.TargetPrice)
		return false
	}
	if e.EntrySatisfied != a.EntrySatisfied {
		fmt.Printf("EntrySatisfied Expect:%v Actual:%v\n", e.EntrySatisfied, a.EntrySatisfied)
		return false
	}
	if e.ExitSatisfied != a.ExitSatisfied {
		fmt.Printf("ExitSatisfied Expect:%v Actual:%v\n", e.ExitSatisfied, a.ExitSatisfied)
		return false
	}
	// Trade
	return true
}

func (b *TestScenario) CheckState(a *atmcs.ATMcs, t *testing.T) {
	now := timeObj
	for _, timestate := range b.TimeStates {
		if timestate.Time.Equal(now) {
			if !timestate.ATMcsState.StatesEqual(a) {
				t.Fatalf("States Not matching")
			}
		}
	}
}

func TestFlow(t *testing.T) {
	LoadTimeLocation()
	log.SetFlags(log.Lshortfile)

	testCases := []string{
		"scene1.json",
		// Add more test case file names here
	}
	settings := []string{
		"testcase1settings.json",
		"testcase1settings.json",
		"testcase1settings.json",
		"testcase1settings.json",
		"testcase1settings.json",
		"testcase1settings.json",
		"testcase1settings.json",
		"testcase1settings.json",
	}
	for i, tc := range testCases {
		fmt.Println("Running testecase file ", tc)

		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Error getting working directory: %v", err)
		}

		data, err := ioutil.ReadFile(filepath.Join(wd, "testcases", "flow", tc))
		if err != nil {
			t.Fatalf("Error reading test case file: %v", err)
		}

		var testCase TestScenario
		err = json.Unmarshal(data, &testCase)
		if err != nil {
			t.Fatalf("Error unmarshalling JSON data: %v", err)
		}

		settingsFileName := settings[i]
		currentFilePath, err := os.Getwd()
		if err != nil {
			t.Fatalf("Error getting current file path: %v", err)
		}
		settingsFilePath := filepath.Join(currentFilePath, "testcases", "entry", settingsFileName)

		timeObj = testCase.CurrentTime.Add(-5 * time.Minute)
		var actualObj executor.ExecutorLike
		timeFunc := func() time.Time {
			timeObj = timeObj.Add(5 * time.Minute)
			return timeObj
		}
		var actualObjATMcs = atmcs.New(settingsFilePath, timeFunc)
		actualObj = actualObjATMcs
		if actualObjATMcs == nil {
			t.Fatalf("ATMcs object init fail")
		}
		broker := NewScenarioBroker(t)
		broker.Expiries = testCase.Expiries
		broker.BidAsks = testCase.OptionDepths
		actualObj.SetBroker(&broker)
		for {
			if actualObj.InTradingWindow() {
				inTrade := actualObj.InTrade()
				if !inTrade && actualObj.IsEntrySatisfied() {
					fmt.Printf("%+v\n", actualObjATMcs)
					// t.SkipNow()
					// actualObj.PaperTrade(actualObj.GetTradeType())
				}
				if inTrade {
					t.SkipNow()
				}
			} else {
				break
			}
			time.Sleep(time.Millisecond * 200)
			testCase.CheckState(actualObjATMcs, t)
		}
	}

}
