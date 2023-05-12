package executor

import "time"

type ExecutorLike interface {
	SetBroker(BrokerLike)
	SetTradeFilePath(string)
	SetSettingsFilesPath(string)
	InTradingWindow() bool
	InTrade() bool
	IsEntrySatisfied() bool
	IsExitSatisfied() bool
	GetEntryMessage() string
	GetExitMessage() string
	IsError() bool
	ReadErrors() []string
	GetSleepDuration() time.Duration
	PaperTrade(TradeType)
	AccountTrade(TradeType)
	GetTradeType() TradeType
	ExitPaper()
	ExitAccount()
}

type Trader struct {
	ID                        string
	HolidaysFilePath          string
	PaperTradeFilePath        string
	AccountTradeFilePath      string
	ExecutorErrorFilePath     string
	SettingsFilePath          string
	BrokerCredentialsFilePath string
	Executor                  ExecutorLike
}
