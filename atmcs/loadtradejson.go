package atmcs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/dragonzurfer/trader/atmcs/trade"
)

func (obj *ATMcs) LoadTradeFromJSON() error {
	fullPath := filepath.Join(obj.Settings.TradeFilePath)

	// Open the file
	file, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read the file into a byte slice
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	// Unmarshal the byte slice into the Trade object
	err = json.Unmarshal(bytes, &obj.Trade)
	if err != nil {
		return err
	}

	return nil
}

func (obj *ATMcs) LoadATMcsTradeToJSON() error {
	// Convert the Trade object to a JSON string
	tradeJSON, err := json.MarshalIndent(obj.Trade, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to convert Trade object to JSON: %w", err)
	}

	// Get the full path of the trade.json file
	fullPath := obj.TradeFilePath

	// Write the JSON string to the file
	err = ioutil.WriteFile(fullPath, tradeJSON, 0644)
	if err != nil {
		return fmt.Errorf("failed to write Trade object to JSON file: %w", err)
	}

	return nil
}

func (obj *ATMcs) InTrade() bool {
	// Define the full path to the JSON file
	fullPath := filepath.Join(obj.TradeFilePath, "trade.json")

	// Load the JSON file
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		log.Printf("Failed to read file %s: %v\n", fullPath, err)
		return obj.Trade.InTrade
	}

	// Unmarshal the JSON into the Trade structure
	var trade trade.Trade
	err = json.Unmarshal(data, &trade)
	if err != nil {
		log.Printf("Failed to unmarshal JSON: %v\n", err)
		return obj.Trade.InTrade
	}

	// Set obj.Trade to the loaded trade data
	obj.Trade = trade

	// Return whether or not a trade is active
	return obj.Trade.InTrade
}
