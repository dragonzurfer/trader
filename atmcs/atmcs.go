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
	ISTLocation    *time.Location
	SignalCPR      cpr.Signal
	EntrySatisfied bool
	ExitSatisfied  bool
	Broker         executor.BrokerLike
	EntryMessage   string
	ExitMessage    string
	SleepDuration  time.Duration
	Trade          trade.Trade
	Settings
	SettingsFilesPath string
	GetCurrentTime    func() time.Time
	Holidays
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
}

func (obj *ATMcs) loadSettingsFromFile() error {
	var settings Settings

	// Read the JSON file
	fileData, err := ioutil.ReadFile(obj.SettingsFilesPath)
	if err != nil {
		return err
	}

	// Unmarshal the JSON data into the Settings struct
	err = json.Unmarshal(fileData, &settings)
	if err != nil {
		return err
	}

	obj.Settings = settings
	return nil
}

func (obj *ATMcs) SetBroker(broker executor.BrokerLike) {
	obj.Broker = broker
}

func (obj *ATMcs) SetTradeFilePath(filepath string) {
	obj.TradeFilePath = filepath
}

func (obj *ATMcs) SetSettingsFilesPath(filepath string) error {
	var err error
	obj.SettingsFilesPath = filepath
	err = obj.loadSettingsFromFile()
	return err
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

func (obj *ATMcs) LoadLocation() error {
	istLocation, err := time.LoadLocation("Asia/Kolkata")
	obj.ISTLocation = istLocation
	return err
}

func (obj *ATMcs) LoadHolidays() error {
	holidaysData, err := ioutil.ReadFile(obj.HolidayDatesFilePath)
	if err != nil {
		log.Println("Error reading holidays file:", err)
		return err
	}

	var holidays Holidays
	err = json.Unmarshal(holidaysData, &holidays)
	if err != nil {
		log.Println("Error unmarshaling holidays data:", err)
		return err
	}
	obj.Holidays = holidays
	return nil
}

func New(settingsFilePath, tradeFilePath string, currentTimeFunc func() time.Time) *ATMcs {
	var obj ATMcs
	if err := obj.SetSettingsFilesPath(settingsFilePath); err != nil {
		log.Println("error loading settings file:", err.Error())
		return nil
	}
	if err := obj.LoadHolidays(); err != nil {
		log.Printf("error loading holidays file from %+v:%v\n", obj.Settings.HolidayDatesFilePath, err.Error())
		return nil
	}
	if err := obj.LoadLocation(); err != nil {
		log.Println("error loading ist location:", err.Error())
		return nil
	}
	obj.SetTradeFilePath(tradeFilePath)
	obj.GetCurrentTime = currentTimeFunc
	obj.EntrySatisfied = false
	obj.ExitSatisfied = false
	return &obj
}
