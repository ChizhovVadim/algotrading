package main

import (
	"encoding/xml"
	"os"

	"github.com/ChizhovVadim/algotrading/internal/candleprovider"
)

type Settings struct {
	SecurityCodes []candleprovider.SecurityCode `xml:"SecurityCode"`
}

func loadSettings(filePath string) (Settings, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return Settings{}, err
	}
	defer file.Close()

	var result Settings
	err = xml.NewDecoder(file).Decode(&result)
	if err != nil {
		return Settings{}, err
	}
	return result, nil
}
