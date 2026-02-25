package main

import (
	"flag"
	"fmt"

	"github.com/ChizhovVadim/algotrading/domain/advisorstatus"
	"github.com/ChizhovVadim/algotrading/domain/model"
	"github.com/ChizhovVadim/algotrading/internal/candlestorage"
	"github.com/ChizhovVadim/algotrading/internal/cli"
	"github.com/ChizhovVadim/algotrading/internal/moex"
)

// Текущая позиция торговой системы.
func statusHandler(args []string) error {
	var (
		advisorName   = "main"
		timeframeName = model.CandleIntervalMinutes5
		securityName  = ""
		count         = 1
	)

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&advisorName, "advisor", advisorName, "")
	flagset.StringVar(&timeframeName, "timeframe", timeframeName, "")
	flagset.StringVar(&securityName, "security", securityName, "")
	flagset.IntVar(&count, "count", count, "")
	flagset.Parse(args)

	if securityName == "" {
		return fmt.Errorf("security required")
	}

	var candleStorage = candlestorage.FromCandleInterval(cli.MapPath("~/TradingData"), timeframeName, moex.Moscow)

	return advisorstatus.ShowStatus(candleStorage, advisorName, securityName, count)
}
