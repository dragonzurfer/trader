package atmcs

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	cpr "github.com/dragonzurfer/strategy/CPR"
	"github.com/dragonzurfer/trader/atmcs/trade"
	"github.com/dragonzurfer/trader/executor"
)

type ATMcs struct {
	Signal         cpr.Signal
	EntrySatisfied bool
	ExitSatisfied  bool
	Broker         executor.BrokerLike
	EntryMessage   string
	ExitMessage    string
	SleepDuration  time.Duration
	Trade          trade.Trade
	Settings
	SettingsFilesPath string
}

type Settings struct {
	TradeFilePath   string  `json:"tradeFilePath"`
	Quantity        int64   `json:"quantity"`
	StrikeDiff      float64 `json:"strikeDiff"`
	MinDaysToExpiry int64   `json:"minDaysToExpiry"`
	Symbol          string  `json:"symbol"`
}

func (obj *ATMcs) loadSettingsFromFile() {
	var settings Settings

	// Read the JSON file
	fileData, err := ioutil.ReadFile(obj.SettingsFilesPath)
	if err != nil {
		log.Printf("Unable to laod setting from %v : %v", obj.SettingsFilesPath, err.Error())
		return
	}

	// Unmarshal the JSON data into the Settings struct
	err = json.Unmarshal(fileData, &settings)
	if err != nil {
		log.Printf("Unable to load setting : %v", err.Error())
		return
	}

	obj.Settings = settings
}

func (obj *ATMcs) SetBroker(broker executor.BrokerLike) {
	obj.Broker = broker
}

func (obj *ATMcs) SetTradeFilePath(filepath string) {
	obj.TradeFilePath = filepath
}

func (obj *ATMcs) SetSettingsFilesPath(filepath string) {
	obj.SettingsFilesPath = filepath
	obj.loadSettingsFromFile()
}

func (obj *ATMcs) InTradingWindow() bool {
	return true
}

func (obj *ATMcs) InTrade() bool {
	return obj.Trade.InTrade
}

func (obj *ATMcs) IsExitSatisfied() bool {
	return true
}
func (obj *ATMcs) GetEntryMessage() string {
	return ""
}
func (obj *ATMcs) GetExitMessage() string {
	return ""
}
func (obj *ATMcs) IsError() bool {
	return false
}
func (obj *ATMcs) ReadErrors() []string {
	return []string{}
}
func (obj *ATMcs) GetSleepDuration() time.Duration {
	return time.Minute
}

func New(settingsFilePath, tradeFilePath string) *ATMcs {
	var obj ATMcs
	obj.SetSettingsFilesPath(settingsFilePath)
	obj.SetTradeFilePath(tradeFilePath)
	return &obj
}
