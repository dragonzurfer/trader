package atmcs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dragonzurfer/trader/executor"
	"github.com/stretchr/testify/assert"
)

func TestExitOnTick(t *testing.T) {

	// You can replace this with actual test settings file path
	currentFilePath, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting current file path: %v", err)
	}
	settingsFilePath := filepath.Join(currentFilePath, "testcases", "entrysatisfy", "testcase1settings.json")
	timeFunc := func() time.Time {
		return time.Now()
	}
	atm := New(settingsFilePath, timeFunc)
	if atm == nil {
		t.Fatalf("ATMcs object init fail")
	}
	atm.SetEntryStates()

	// Set the trade type and entry price
	atm.Trade.InTrade = true
	atm.Trade.TradeType = executor.Buy
	atm.Trade.EntryPrice = 100.0

	// Test the case where stop loss should be hit
	atm.Trade.StopLossPrice = 90.0

	slHitChannel := atm.GetStopLossHitChan()
	targetHitChannel := atm.GetTargetHitChan()

	go func() {
		time.Sleep(time.Second)
		atm.ExitOnTick(85.0)
	}()

	select {
	case _, ok := <-slHitChannel:
		if !ok {
			t.Fatal("StopLossHitChan was closed")
		}
		assert.True(t, atm.ExitSatisfied, "Expected ExitSatisfied to be true")
	case <-time.After(time.Second * 5):
		t.Fatal("Timeout waiting for StopLossHitChan")
	}

	// Reset ExitSatisfied
	atm.ExitSatisfied = false

	// Test the case where target should be hit
	atm.Trade.TargetPrice = 120.0

	go func() {
		time.Sleep(time.Second)
		atm.ExitOnTick(125.0)
	}()

	select {
	case <-targetHitChannel:
		assert.True(t, atm.ExitSatisfied, "Expected ExitSatisfied to be true")
	case <-time.After(time.Second * 5):
		t.Fatal("Timeout waiting for TargetHitChan")
	}
}
