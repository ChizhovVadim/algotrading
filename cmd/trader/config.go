package main

import (
	"encoding/xml"
	"os"

	"github.com/ChizhovVadim/algotrading/internal/traderapp"
)

func loadConfig(filePath string) (traderapp.Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return traderapp.Config{}, err
	}
	defer file.Close()

	var result traderapp.Config
	err = xml.NewDecoder(file).Decode(&result)
	if err != nil {
		return traderapp.Config{}, err
	}
	return result, nil
}
