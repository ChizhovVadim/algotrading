package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/model"
	"github.com/ChizhovVadim/algotrading/internal/candleprovider"
	"github.com/ChizhovVadim/algotrading/internal/cli"
)

func testDownloadHandler(args []string) error {
	var (
		providerName  string
		timeframeName = model.CandleIntervalMinutes5
		securityName  string
		startDate     = cli.DateValue{Date: time.Now().AddDate(0, 0, -5)}
		finishDate    = cli.DateValue{Date: time.Now()}
	)

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&providerName, "provider", providerName, "")
	flagset.StringVar(&timeframeName, "timeframe", timeframeName, "")
	flagset.StringVar(&securityName, "security", securityName, "")
	flagset.Var(&startDate, "start", "")
	flagset.Var(&finishDate, "finish", "")
	flagset.Parse(args)

	if securityName == "" {
		return fmt.Errorf("security required")
	}

	settings, err := loadSettings("seccodes.xml")
	if err != nil {
		return err
	}

	var candleProvider = candleprovider.NewMultyProvider(settings.SecurityCodes, timeframeName)
	err = candleProvider.AddSource(providerName)
	if err != nil {
		return err
	}

	candles, err := candleProvider.Load(securityName, startDate.Date, finishDate.Date)
	if err != nil {
		return err
	}

	fmt.Println("Downloaded",
		"size", len(candles))
	for _, candle := range compactCandles(candles, 10) {
		fmt.Println(candle)
	}

	return nil
}

func compactCandles(source []model.Candle, size int) []model.Candle {
	if len(source) <= size {
		return source
	}
	var result []model.Candle
	result = append(result, source[:size/2]...)
	result = append(result, source[len(source)-size/2:]...)
	return result
}
