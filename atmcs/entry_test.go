package atmcs_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"trader/atmcs"

	cpr "github.com/dragonzurfer/strategy/CPR"
	"github.com/dragonzurfer/trader/executor"
	"github.com/stretchr/testify/assert"
)

var ISTLocation *time.Location

func LoadTimeLocation() {
	ISTLocation, _ = time.LoadLocation("Asia/Kolkata")
}

type Settings struct {
	TradeFilePath   string  `json:"tradeFilePath"`
	Quantity        int64   `json:"quantity"`
	StrikeDiff      float64 `json:"strikeDiff"`
	MinDaysToExpiry int64   `json:"minDaysToExpiry"`
	Symbol          string  `json:"symbol"`
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
	TradeType executor.TradeType `json:"trade_type"`
	Quantity  int64              `json:"quantity"`
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

func (b *TestBroker) GetCandles(string) ([]executor.CandleLike, error) {
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

func (b *TestBroker) GetCandlesOption(float64, time.Time, executor.OptionType) ([]executor.CandleLike, error) {
	return nil, nil
}

func TestPaperTrade(t *testing.T) {
	testCases := []string{
		"testcase1.json",

		// Add more test case file names here
	}
	settings := []string{
		"testcase1settings.json",
	}

	for i, tc := range testCases {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Error getting working directory: %v", err)
		}

		data, err := ioutil.ReadFile(filepath.Join(wd, "testcases", tc))
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
		settingsFilePath := filepath.Join(currentFilePath, "testcases", settingsFileName)
		//create obj
		LoadTimeLocation()
		acutalObj := atmcs.New(settingsFilePath, currentFilePath, func() time.Time { return testCase.CurrentTime })
		var broker TestBroker
		broker.Expiries = testCase.OptionExpiries
		broker.BidAsks = testCase.OptionDepths
		broker.LTP = testCase.LTP

		acutalObj.SetBroker(&broker)
		acutalObj.PaperTrade(executor.Buy)
		// see if atmcs.Trade == TestCase.ExpectedTrade
		fmt.Println(acutalObj.GetCurrentTime(), "hello")
		fmt.Printf("\nExpected:%+v\nActual:%+v", testCase.ExpectedTrade.EntryPositions, acutalObj.Trade.EntryPositions)
		if len(acutalObj.Trade.EntryPositions) != len(testCase.ExpectedTrade.EntryPositions) {
			t.Fatalf("\nActual:%+v\nExpected:%+v", len(acutalObj.Trade.EntryPositions), len(testCase.ExpectedTrade.EntryPositions))
		}
		for i := 0; i < len(acutalObj.Trade.EntryPositions); i += 1 {
			if acutalObj.Trade.EntryPositions[i].GetExpiry() != testCase.ExpectedTrade.EntryPositions[i].GetExpiry() {
				t.Fatalf("\nActual:%+v\nExpected:%+v", acutalObj.Trade.EntryPositions[i].GetExpiry(), testCase.ExpectedTrade.EntryPositions[i].GetExpiry())
				assert.Equal(t, acutalObj.Trade.EntryPositions, testCase.ExpectedTrade.EntryPositions)
			}
		}

	}
}
