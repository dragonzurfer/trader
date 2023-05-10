package atmcs_test

// func TestIsEntrySatisfied(t *testing.T) {
// 	LoadTimeLocation()
// 	log.SetFlags(log.Lshortfile)
// 	historicalTime := []string{
// 		"2023-05-01T00:00:00+05:30",
// 	}
// 	settings := []string{
// 		"testcase1settings.json",
// 	}

// 	for i, timeString := range historicalTime {
// 		fmt.Println("Running for time: ", timeString)
// 		timeObj := // convert timeString from string to time.Time object

// 		settingsFileName := settings[0]
// 		currentFilePath, err := os.Getwd()
// 		if err != nil {
// 			t.Fatalf("Error getting current file path: %v", err)
// 		}
// 		settingsFilePath := filepath.Join(currentFilePath, "testcases", settingsFileName)
// 		actualObj := atmcs.New(settingsFilePath, currentFilePath, func() time.Time { return timeObj })
// 		var broker TestBroker
// 		broker.Expiries = testCase.OptionExpiries
// 		broker.BidAsks = testCase.OptionDepths
// 		broker.LTP = testCase.LTP

// 		actualObj.SetBroker(&broker)
// 		if testCase.Signal.Signal == cpr.Buy {
// 			actualObj.PaperTrade(executor.Buy)
// 		} else {
// 			actualObj.PaperTrade(executor.Sell)
// 		}
// 		// see if atmcs.Trade == TestCase.ExpectedTrade
// 		fmt.Printf("\nExpected:%+v\nActual:%+v\n", testCase.ExpectedTrade.EntryPositions, actualObj.Trade.EntryPositions)
// 		if len(actualObj.Trade.EntryPositions) != len(testCase.ExpectedTrade.EntryPositions) {
// 			t.Fatalf("\nActual:%+v\nExpected:%+v\n", len(actualObj.Trade.EntryPositions), len(testCase.ExpectedTrade.EntryPositions))
// 		}
// 		for i := 0; i < len(actualObj.Trade.EntryPositions); i++ {
// 			entryActual := actualObj.Trade.EntryPositions[i]
// 			entryExpected := testCase.ExpectedTrade.EntryPositions[i]

// 			// Compare Expiry field
// 			if !entryActual.GetExpiry().Equal(entryExpected.GetExpiry()) {
// 				t.Fatalf("\nActual Expiry: %+v\nExpected Expiry: %+v\n", entryActual.GetExpiry(), entryExpected.GetExpiry())
// 			}

// 			// Compare Strike field
// 			if entryActual.GetStrike() != entryExpected.GetStrike() {
// 				t.Fatalf("\nActual Strike: %f\nExpected Strike: %f\n", entryActual.GetStrike(), entryExpected.GetStrike())
// 			}

// 			// Compare Type field
// 			if entryActual.GetOptionType() != entryExpected.GetOptionType() {
// 				t.Fatalf("\nActual Type: %s\nExpected Type: %s\n", entryActual.GetOptionType(), entryExpected.GetOptionType())
// 			}

// 			// Compare Symbol field
// 			if entryActual.GetOptionSymbol() != entryExpected.GetOptionSymbol() {
// 				t.Fatalf("\nActual Symbol: %s\nExpected Symbol: %s\n", entryActual.GetOptionSymbol(), entryExpected.GetOptionSymbol())
// 			}

// 			// Compare UnderlyingSymbol field
// 			if entryActual.GetUnderlyingSymbol() != entryExpected.GetUnderlyingSymbol() {
// 				t.Fatalf("\nActual UnderlyingSymbol: %s\nExpected UnderlyingSymbol: %s\n", entryActual.GetUnderlyingSymbol(), entryExpected.GetUnderlyingSymbol())
// 			}

// 			// Compare Price field
// 			if entryActual.GetPrice() != entryExpected.GetPrice() {
// 				t.Fatalf("\nActual Price: %f\nExpected Price: %f\n", entryActual.GetPrice(), entryExpected.GetPrice())
// 			}

// 			// Compare TradeType field
// 			if entryActual.GetTradeType() != entryExpected.GetTradeType() {
// 				t.Fatalf("\nActual TradeType: %s\nExpected TradeType: %s\n", entryActual.GetTradeType(), entryExpected.GetTradeType())
// 			}

// 			// Compare Quantity field
// 			if entryActual.GetQuantity() != entryExpected.GetQuantity() {
// 				t.Fatalf("\nActual Quantity: %d\nExpected Quantity: %d\n", entryActual.GetQuantity(), entryExpected.GetQuantity())
// 			}
// 		}

// 	}
// }
