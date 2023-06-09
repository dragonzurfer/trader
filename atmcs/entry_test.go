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

	"github.com/dragonzurfer/trader/atmcs"

	cpr "github.com/dragonzurfer/strategy/CPR"
	"github.com/dragonzurfer/trader/executor"
)

var ISTLocation *time.Location

func LoadTimeLocation() {
	ISTLocation, _ = time.LoadLocation("Asia/Kolkata")
}

type Holidays struct {
	HolidayDates []string `json:"holiday_dates"`
}

type Settings struct {
	HolidayDatesFilePath string  `json:"holidays_file_path"`
	MinTrailPercent      float64 `json:"min_trail_percent"`
	MinTargetPercent     float64 `json:"min_target_percent"`
	MinStopLossPercent   float64 `json:"min_sl_percent"`
	TradeFilePath        string  `json:"tradeFilePath"`
	Quantity             int64   `json:"quantity"`
	StrikeDiff           float64 `json:"strikeDiff"`
	MinDaysToExpiry      int64   `json:"minDaysToExpiry"`
	Symbol               string  `json:"symbol"`
	TickSize             float64 `json:"tick_size"`
}

type MarketDepth struct {
	Price       float64 `json:"price"`
	Quantity    int64   `json:"quantity"`
	NumOfOrders int64   `json:"num_of_orders"`
}

type BidAsk struct {
	Bids []MarketDepth `json:"bids"`
	Asks []MarketDepth `json:"asks"`
}

func (md MarketDepth) GetPrice() float64              { return md.Price }
func (md MarketDepth) GetQuantity() int64             { return md.Quantity }
func (md MarketDepth) GetNumOfOrders() int64          { return md.NumOfOrders }
func (ba BidAsk) GetBids() []executor.MarketDepthLike { return toMarketDepthLikeSlice(ba.Bids) }
func (ba BidAsk) GetAsks() []executor.MarketDepthLike { return toMarketDepthLikeSlice(ba.Asks) }

func toMarketDepthLikeSlice(marketDepths []MarketDepth) []executor.MarketDepthLike {
	depthLikes := make([]executor.MarketDepthLike, len(marketDepths))
	for i, depth := range marketDepths {
		depthLikes[i] = depth
	}
	return depthLikes
}

type Option struct {
	Expiry           time.Time           `json:"expiry"`
	Strike           float64             `json:"strike"`
	Type             executor.OptionType `json:"type"`
	Symbol           string              `json:"symbol"`
	UnderlyingSymbol string              `json:"underlying_symbol"`
}

type OptionPosition struct {
	Option
	Price     float64            `json:"price"`
	TradeType executor.TradeType `json:"TradeType"`
	Quantity  int64              `json:"Quantity"`
}

func (op OptionPosition) GetTradeType() executor.TradeType { return op.TradeType }
func (op OptionPosition) GetPrice() float64                { return op.Price }
func (op OptionPosition) GetQuantity() int64               { return op.Quantity }

func (o Option) GetExpiry() time.Time               { return o.Expiry }
func (o Option) GetStrike() float64                 { return o.Strike }
func (o Option) GetOptionType() executor.OptionType { return o.Type }
func (o Option) GetOptionSymbol() string            { return o.Symbol }
func (o Option) GetUnderlyingSymbol() string        { return o.UnderlyingSymbol }

type ExpectedTrade struct {
	InTrade            bool             `json:"in_trade"`
	TrailStopLossPrice float64          `json:"trail_sl"`
	EntryPositions     []OptionPosition `json:"entry_positions"`
	ExitPositions      []OptionPosition `json:"exit_positions"`
	TimeOfEntry        time.Time        `json:"time_of_entry"`
	TimeOfExit         time.Time        `json:"time_of_exit"`
	StopLossPrice      float64          `json:"stop_loss_price"`
	TargetPrice        float64          `json:"target_price"`
}

type TestOptionDepth struct {
	Option Option `json:"option"`
	BidAsk BidAsk `json:"bid_ask"`
}

type TestCase struct {
	CurrentTime    time.Time            `json:"current_time"`
	LTP            float64              `json:"ltp"`
	Signal         cpr.Signal           `json:"signal"`
	ExpectedTrade  ExpectedTrade        `json:"expected_trade"`
	OptionExpiries []executor.Expiry    `json:"option_chains"`
	OptionDepths   map[time.Time]BidAsk `json:"option_depths"`
}

type TestBroker struct {
	LTP      float64
	Expiries []executor.Expiry
	BidAsks  map[time.Time]BidAsk
}

func (b *TestBroker) SetCredentialsFilePath(string) {}

func (b *TestBroker) GetLTP(string) (float64, error) {
	return b.LTP, nil
}

func (b *TestBroker) GetMarketDepth(string) (executor.BidAskLike, error) {
	return b.BidAsks[time.Now()], nil
}

func (b *TestBroker) GetCandles(string, time.Time, time.Time, executor.TimeFrame) ([]executor.CandleLike, error) {
	return nil, nil
}

func (b *TestBroker) GetOptionExpiries(string) ([]executor.Expiry, error) {
	return b.Expiries, nil
}

func (b *TestBroker) GetMarketDepthOption(strike float64, optExpiry time.Time, optType executor.OptionType) (executor.BidAskLike, error) {
	for expiry, bidask := range b.BidAsks {
		if expiry == optExpiry {
			return bidask, nil
		}
	}
	return nil, errors.New("API Could not find expiry date")
}

func (b *TestBroker) GetCandlesOption(float64, time.Time, executor.OptionType, time.Time, time.Time) ([]executor.CandleLike, error) {
	return nil, nil
}

func TestPaperTrade(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	testCases := []string{
		"testcase1.json",
		"testcase2.json",
		"testcase3.json",
		"testcase4.json",
		"testcase5.json",
		"testcase6.json",
		"testcase7.json",
		"testcase8.json",
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

		data, err := ioutil.ReadFile(filepath.Join(wd, "testcases", "entry", tc))
		if err != nil {
			t.Fatalf("Error reading test case file: %v", err)
		}

		var testCase TestCase
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
		//create obj
		LoadTimeLocation()
		actualObj := atmcs.New(settingsFilePath, func() time.Time { return testCase.CurrentTime })
		if actualObj == nil {
			t.Fatal("Failed to create atmcs object")
		}
		var broker TestBroker
		broker.Expiries = testCase.OptionExpiries
		broker.BidAsks = testCase.OptionDepths
		broker.LTP = testCase.LTP

		actualObj.SetBroker(&broker)
		if testCase.Signal.Signal == cpr.Buy {
			actualObj.PaperTrade(executor.Buy)
		} else {
			actualObj.PaperTrade(executor.Sell)
		}
		// see if atmcs.Trade == TestCase.ExpectedTrade
		fmt.Printf("\nExpected:%+v\nActual:%+v\n", testCase.ExpectedTrade.EntryPositions, actualObj.Trade.EntryPositions)
		if len(actualObj.Trade.EntryPositions) != len(testCase.ExpectedTrade.EntryPositions) {
			t.Fatalf("\nActual:%+v\nExpected:%+v\n", len(actualObj.Trade.EntryPositions), len(testCase.ExpectedTrade.EntryPositions))
		}
		for i := 0; i < len(actualObj.Trade.EntryPositions); i++ {
			entryActual := actualObj.Trade.EntryPositions[i]
			entryExpected := testCase.ExpectedTrade.EntryPositions[i]

			if actualObj.Trade.TimeOfEntry != testCase.CurrentTime {
				t.Fatalf("Actual:%v\nExpected:%v\n", actualObj.Trade.TimeOfEntry, testCase.CurrentTime)
			}

			// Compare Expiry field
			if !entryActual.GetExpiry().Equal(entryExpected.GetExpiry()) {
				t.Fatalf("\nActual Expiry: %+v\nExpected Expiry: %+v\n", entryActual.GetExpiry(), entryExpected.GetExpiry())
			}

			// Compare Strike field
			if entryActual.GetStrike() != entryExpected.GetStrike() {
				t.Fatalf("\nActual Strike: %f\nExpected Strike: %f\n", entryActual.GetStrike(), entryExpected.GetStrike())
			}

			// Compare Type field
			if entryActual.GetOptionType() != entryExpected.GetOptionType() {
				t.Fatalf("\nActual Type: %s\nExpected Type: %s\n", entryActual.GetOptionType(), entryExpected.GetOptionType())
			}

			// Compare Symbol field
			if entryActual.GetOptionSymbol() != entryExpected.GetOptionSymbol() {
				t.Fatalf("\nActual Symbol: %s\nExpected Symbol: %s\n", entryActual.GetOptionSymbol(), entryExpected.GetOptionSymbol())
			}

			// Compare UnderlyingSymbol field
			if entryActual.GetUnderlyingSymbol() != entryExpected.GetUnderlyingSymbol() {
				t.Fatalf("\nActual UnderlyingSymbol: %s\nExpected UnderlyingSymbol: %s\n", entryActual.GetUnderlyingSymbol(), entryExpected.GetUnderlyingSymbol())
			}

			// Compare Price field
			if entryActual.GetPrice() != entryExpected.GetPrice() {
				t.Fatalf("\nActual Price: %f\nExpected Price: %f\n", entryActual.GetPrice(), entryExpected.GetPrice())
			}

			// Compare TradeType field
			if entryActual.GetTradeType() != entryExpected.GetTradeType() {
				t.Fatalf("\nActual TradeType: %s\nExpected TradeType: %s\n", entryActual.GetTradeType(), entryExpected.GetTradeType())
			}

			// Compare Quantity field
			if entryActual.GetQuantity() != entryExpected.GetQuantity() {
				t.Fatalf("\nActual Quantity: %d\nExpected Quantity: %d\n", entryActual.GetQuantity(), entryExpected.GetQuantity())
			}
		}

		if !actualObj.Trade.TimeOfEntry.Equal(testCase.CurrentTime) {
			t.Fatalf("\nActual timeofentry:%v Expected:%v", actualObj.Trade.TimeOfEntry, testCase.CurrentTime)
		}
		actualObj.ExitPaper()
		fmt.Printf("%+v\n", actualObj.Trade)

	}
}
