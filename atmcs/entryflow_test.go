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
	cpr "github.com/dragonzurfer/strategy/CPR"
	"github.com/dragonzurfer/trader/atmcs"
	"github.com/dragonzurfer/trader/executor"
	"github.com/stretchr/testify/assert"
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
	Time   time.Time `json:"time"`
	Price  float64   `json:"price"`
	IsSent bool
}

type TimeState struct {
	Time       time.Time  `json:"time"`
	ATMcsState SceneATMcs `json:"atmcs_state"`
}

type TestScenario struct {
	CurrentTime      time.Time            `json:"current_time"`
	TickPrices       []TickPrice          `json:"tick_prices"`
	TimeStates       []TimeState          `json:"time_states"`
	Expiries         []executor.Expiry    `json:"option_chains"`
	OptionDepths     map[time.Time]BidAsk `json:"option_depths"`
	TimeObj          *time.Time
	SlHitChannel     <-chan bool
	TargetHitChannel <-chan bool
	TrailChannel     <-chan bool
	T                *testing.T
	atm              executor.ExecutorLike
	atmRef           *atmcs.ATMcs
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
	// fmt.Println(startTime, ":", endTime)
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
	if a == nil {
		log.Println("Getting nil for actual ATMcs")
		return false
	}
	if e == nil {
		log.Println("Getting nil for expected ATMcs")
		return false
	}

	if e.EntrySatisfied != a.EntrySatisfied {
		log.Printf("EntrySatisfied Expect:%v Actual:%v\n", e.EntrySatisfied, a.EntrySatisfied)
		return false
	}
	if e.ExitSatisfied != a.ExitSatisfied {
		log.Printf("ExitSatisfied Expect:%v Actual:%v\n", e.ExitSatisfied, a.ExitSatisfied)
		return false
	}
	// Trade
	et := e.Trade
	at := a.Trade

	if et.InTrade != at.InTrade {
		log.Printf("InTrade Expect:%v Actual:%v\n", et.InTrade, at.InTrade)
		return false
	}
	if !et.TimeOfEntry.Equal(at.TimeOfEntry) {
		log.Printf("TimeOfEntry Expect:%v Actual:%v\n", et.TimeOfEntry, at.TimeOfEntry)
		return false
	}
	if !et.TimeOfExit.Equal(at.TimeOfExit) {
		log.Printf("TimeOfExit Expect:%v Actual:%v\n", et.TimeOfExit, at.TimeOfExit)
		return false
	}

	if et.TradeType != at.TradeType {
		log.Printf("TradeType Expect:%v Actual:%v\n", et.TradeType, at.TradeType)
		return false
	}
	if et.IsMinTrailHit != at.IsMinTrailHit {
		log.Printf("IsMinTrailHit Expect:%v Actual:%v\n", et.IsMinTrailHit, at.IsMinTrailHit)
		return false
	}
	if et.IsStopLossHit != at.IsStopLossHit {
		log.Printf("IsStopLossHit Expect:%v Actual:%v\n", et.IsStopLossHit, at.IsStopLossHit)
		return false
	}
	if e.EntrySatisfied {
		if at.EntryPositions == nil {
			log.Printf("Entry positions empty Expected:%+v\n", et.ExitPositions)
			return false
		}
	}
	if len(at.EntryPositions) != len(et.EntryPositions) {
		log.Printf("Length of entries Actual:%v Expected:%v", len(at.EntryPositions), len(et.EntryPositions))
		return false
	}

	n := len(at.EntryPositions)
	apos := at.EntryPositions
	epos := et.EntryPositions
	for i := 0; i < n; i += 1 {
		if apos[i].Option.Strike != epos[i].Option.Strike {
			log.Printf("Option mismatch\nExpect:%+v\nActual:%+v\n", epos[i].Option, apos[i].Option)
			return false
		}
		if apos[i].Option.Type != epos[i].Option.Type {
			log.Printf("Option mismatch\nExpect:%+v\nActual:%+v\n", epos[i].Option, apos[i].Option)
			return false
		}
		if apos[i].Option.UnderlyingSymbol != epos[i].Option.UnderlyingSymbol {
			log.Printf("Option mismatch\nExpect:%+v\nActual:%+v\n", epos[i].Option, apos[i].Option)
			return false
		}
		if !apos[i].Option.Expiry.Equal(epos[i].Option.Expiry) {
			log.Printf("Option mismatch\nExpect:%+v\nActual:%+v\n", epos[i].Option, apos[i].Option)
			return false
		}

		if apos[i].Price != epos[i].Price {
			log.Printf("Price Expected:%v Actual:%v\n", epos[i].Price, apos[i].Price)
			return false
		}
		if apos[i].TradeType != epos[i].TradeType {
			log.Printf("Type Expected:%+v Actual:%v\n", epos[i], apos[i].TradeType)
			return false
		}
		if apos[i].Quantity != epos[i].Quantity {
			log.Printf("Quantity Expected:%v Actual:%v\n", epos[i].Quantity, apos[i].Quantity)
			return false
		}
	}
	if e.ExitSatisfied {
		if at.ExitPositions == nil {
			log.Printf("Exit positions empty Expected:%+v\n", et.ExitPositions)
			return false
		}
	}
	if len(at.ExitPositions) != len(et.ExitPositions) {
		log.Printf("Length of entries Actual:%v Expected:%v\n", len(at.ExitPositions), len(et.ExitPositions))
		return false
	}
	n = len(at.ExitPositions)
	apos = at.ExitPositions
	epos = et.ExitPositions
	for i := 0; i < n; i += 1 {
		if apos[i].Option.Strike != epos[i].Option.Strike {
			log.Printf("Option mismatch\nExpect:%+v\nActual:%+v\n", epos[i].Option, apos[i].Option)
			return false
		}
		if apos[i].Option.Type != epos[i].Option.Type {
			log.Printf("Option mismatch\nExpect:%+v\nActual:%+v\n", epos[i].Option, apos[i].Option)
			return false
		}
		if apos[i].Option.UnderlyingSymbol != epos[i].Option.UnderlyingSymbol {
			log.Printf("Option mismatch\nExpect:%+v\nActual:%+v\n", epos[i].Option, apos[i].Option)
			return false
		}
		if !apos[i].Option.Expiry.Equal(epos[i].Option.Expiry) {
			log.Printf("Option mismatch\nExpect:%+v\nActual:%+v\n", epos[i].Option, apos[i].Option)
			return false
		}

		if apos[i].Price != epos[i].Price {
			log.Printf("Price Expected:%v Actual:%v\n", epos[i].Price, apos[i].Price)
			return false
		}
		if apos[i].TradeType != epos[i].TradeType {
			log.Printf("Price Expected:%v Actual:%v\n", epos[i].TradeType, apos[i].TradeType)
			return false
		}
		if apos[i].Quantity != epos[i].Quantity {
			log.Printf("Price Expected:%v Actual:%v\n", epos[i].Quantity, apos[i].Quantity)
			return false
		}
	}
	return true
}

func (b *TestScenario) CheckState(a *atmcs.ATMcs, t *testing.T) {
	now := b.TimeObj
	for _, timestate := range b.TimeStates {
		if timestate.Time.Equal(*now) {
			fmt.Println("found state for time:", now)
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
		"scene5.json",
		// "2023-05-18.json",
		// Add more test case file names here
	}
	settings := []string{
		"scene_settings.json",
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
		settingsFilePath := filepath.Join(currentFilePath, "testcases", "flow", settingsFileName)

		timeObj = testCase.CurrentTime.Add(-5 * time.Minute)
		var atm executor.ExecutorLike
		timeFunc := func() time.Time {
			return timeObj
		}
		var atmRef = atmcs.New(settingsFilePath, timeFunc)
		atm = atmRef
		if atmRef == nil {
			t.Fatalf("ATMcs object init fail")
		}
		broker := NewScenarioBroker(t)
		broker.Expiries = testCase.Expiries
		broker.BidAsks = testCase.OptionDepths
		atm.SetBroker(&broker)
		testCase.TimeObj = &timeObj
		testCase.atm = atm
		testCase.atmRef = atmRef
		timeObj = timeObj.Add(5 * time.Minute)
		for {
			if atm.InTradingWindow() {
				timeObj = timeObj.Add(5 * time.Minute)
				inTrade := atm.InTrade()
				if !inTrade && atm.IsEntrySatisfied() {
					atm.PaperTrade(atm.GetTradeType())
					err := atm.LogTrade()
					assert.Nil(t, err, "Expected err to be nil")
					fmt.Println(atm.GetEntryMessage())
				}
				if inTrade {
					// t.SkipNow()
					slHitChannel := atm.GetStopLossHitChan()
					targetHitChannel := atm.GetTargetHitChan()
					trailChannel := atmRef.GetTrailChan()
					testCase.TrailChannel = trailChannel
					testCase.SlHitChannel = slHitChannel
					testCase.TargetHitChannel = targetHitChannel
					fmt.Println("listening to channels")
					testCase.listenToChannels()
					err := atm.LogTrade()
					assert.Nil(t, err, "Expected err to be nil")
					fmt.Println(atm.GetExitMessage())
				}
			} else {
				if atm.InTrade() {
					t.Fatalf("Shouldn't be here")
					testCase.atm.ExitPaper()
					err := atm.LogTrade()
					assert.Nil(t, err, "Expected err to be nil")
					fmt.Println(atm.GetExitMessage())
				}
				break
			}
			time.Sleep(time.Millisecond * 200)
			testCase.CheckState(atmRef, t)
			testCase.SetModuloDurationZero()
		}
	}

}

func (scene *TestScenario) SetModuloDurationZero() {
	now := scene.TimeObj
	marketOpenTime := time.Date(now.Year(), now.Month(), now.Day(), 9, 15, 0, 0, ISTLocation)
	duration := scene.atm.GetSleepDuration()
	for {
		durationSinceMarketOpen := now.Sub(marketOpenTime)
		if int(durationSinceMarketOpen.Seconds())%int(duration.Seconds()) == 0 {
			break
		}
		*now = now.Add(time.Second)
	}
	scene.TimeObj = now
}

func (scene *TestScenario) listenToChannels() {
	defer fmt.Println("returning from listenToChannels")
	marketCloseChan := make(chan bool)
	go func() {
		defer fmt.Println("done ticking")
		for {

			// scene.CheckState(scene.atmRef, scene.T)
			*scene.TimeObj = scene.TimeObj.Add(time.Minute)
			fmt.Println("time is :", scene.TimeObj)
			for i, tickprice := range scene.TickPrices {
				if scene.TimeObj.After(tickprice.Time) && !tickprice.IsSent {
					scene.TickPrices[i].IsSent = true
					fmt.Println("sending tick:", tickprice.Price)
					go scene.atm.ExitOnTick(tickprice.Price)
					time.Sleep(time.Millisecond * 200)
				}
			}
			marketCloseTime := time.Date(scene.TimeObj.Year(), scene.TimeObj.Month(), scene.TimeObj.Day(), 15, 15, 0, 0, ISTLocation)
			if scene.atmRef.ExitSatisfied {
				fmt.Println("exit done")
				return
			}
			if scene.TimeObj.After(marketCloseTime) {
				fmt.Println("market closed")
				marketCloseChan <- true
				return
			}
		}
	}()

	for {
		select {
		case slHit, ok := <-scene.SlHitChannel:
			if ok {
				if slHit {
					fmt.Println("Stop Loss hit!")
					// scene.atm.ExitPaper()
					return
				} else {
					fmt.Println("SL not hit")
					return
				}
			} else {
				scene.T.Fatal("sl Channel closed")
				return
			}
		case targetHit, ok := <-scene.TargetHitChannel:
			if ok {
				if targetHit {
					fmt.Println("Target hit!")
					// scene.atm.ExitPaper()
					return
				} else {
					fmt.Println("target not hit")
					return
				}
			} else {
				scene.T.Fatal("target Channel closed")

				return
			}
		case trail, ok := <-scene.TrailChannel:
			if ok {
				if trail {
					fmt.Println("Trail to cost")
				}
			} else {
				scene.T.Fatal("trail Channel closed")
				return
			}
		case <-marketCloseChan:
			fmt.Println("Market about to close, exiting.")
			scene.atm.ExitPaper()
			return
		}
	}
}
