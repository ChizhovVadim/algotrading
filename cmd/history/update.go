package main

import (
	"flag"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ChizhovVadim/algotrading/domain/candleupdate"
	"github.com/ChizhovVadim/algotrading/domain/model"
	"github.com/ChizhovVadim/algotrading/internal/candleprovider"
	"github.com/ChizhovVadim/algotrading/internal/candlestorage"
	"github.com/ChizhovVadim/algotrading/internal/cli"
	"github.com/ChizhovVadim/algotrading/internal/moex"
)

// Скачивание и сохранение новых исторических баров.
func updateHandler(args []string) error {
	var (
		providerName  string
		timeframeName = model.CandleIntervalMinutes5
		securityName  string
		maxDays       = 30
	)

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&providerName, "provider", providerName, "")
	flagset.StringVar(&timeframeName, "timeframe", timeframeName, "")
	flagset.StringVar(&securityName, "security", securityName, "")
	flagset.IntVar(&maxDays, "maxdays", maxDays, "")
	flagset.Parse(args)

	if securityName == "" {
		return fmt.Errorf("security required")
	}

	var securityCodes = strings.Split(securityName, ",")
	var providerNames = strings.Split(providerName, ",")

	settings, err := loadSettings("seccodes.xml")
	if err != nil {
		return err
	}

	var candleStorage = candlestorage.FromCandleInterval(cli.MapPath("~/TradingData"), timeframeName, moex.Moscow)

	var candleProvider = candleprovider.NewMultyProvider(settings.SecurityCodes, timeframeName)
	for _, providerName := range providerNames {
		var err = candleProvider.AddSource(providerName)
		if err != nil {
			return err
		}
	}

	var updateService = candleupdate.New(slog.Default(), candleStorage, candleProvider, maxDays)
	return updateService.UpdateGroup(securityCodes)
}
