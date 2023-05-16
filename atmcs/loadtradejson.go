package atmcs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func (obj *ATMcs) LoadFromJSON() error {
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

func (obj *ATMcs) LogTrade() error {
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
	return obj.Trade.InTrade
}
