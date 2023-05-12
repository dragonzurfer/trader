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
	TickSize             float64 `json:"tick_size"`
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

func (obj *ATMcs) SetSettingsFilesPath(filepath string) {
	obj.SettingsFilesPath = filepath
}

func (obj *ATMcs) InTradingWindow() bool {
	return true
}

func (obj *ATMcs) InTrade() bool {
	return obj.Trade.InTrade
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

func (obj *ATMcs) AccountTrade(executor.TradeType) {

}

func (obj *ATMcs) GetTradeType() executor.TradeType {
	return obj.Trade.TradeType
}

func New(settingsFilePath, tradeFilePath string, currentTimeFunc func() time.Time) *ATMcs {
	var obj ATMcs
	obj.SetSettingsFilesPath(settingsFilePath)

	if err := obj.loadSettingsFromFile(); err != nil {
		log.Println("SetSettingsFIlesPath(string) failed to load settings from", obj.SettingsFilesPath)
		return nil
	}
	if err := obj.LoadHolidays(); err != nil {
		log.Printf("LoadHolidays() file from %+v:%v\n", obj.Settings.HolidayDatesFilePath, err.Error())
		return nil
	}
	if err := obj.LoadLocation(); err != nil {
		log.Println("error loading ist location:", err.Error())
		return nil
	}
	if obj.Settings.TickSize <= 0 {
		log.Println("error loading tick size, cannot be <= 0")
		return nil
	}
	if obj.Settings.MinTargetPercent < obj.Settings.MinTrailPercent {
		log.Println("min target cananot be lesser than min trail in settings")
		return nil
	}
	obj.SetTradeFilePath(tradeFilePath)
	obj.GetCurrentTime = currentTimeFunc
	obj.EntrySatisfied = false
	obj.ExitSatisfied = false
	return &obj
}
