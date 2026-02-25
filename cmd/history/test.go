package main

import (
	"flag"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/historyreport"
	"github.com/ChizhovVadim/algotrading/domain/model"
	"github.com/ChizhovVadim/algotrading/internal/candlestorage"
	"github.com/ChizhovVadim/algotrading/internal/cli"
	"github.com/ChizhovVadim/algotrading/internal/moex"
)

// Проверяем торговую систему на исторических данных.
func testHandler(args []string) error {
	var today = time.Now()
	var r historyreport.ReportRequest

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&r.AdvisorName, "advisor", "main", "")
	flagset.StringVar(&r.TimeframeName, "timeframe", model.CandleIntervalMinutes5, "")
	flagset.StringVar(&r.SecurityName, "security", "Si", "")
	flagset.Float64Var(&r.Lever, "lever", 0, "")
	flagset.Float64Var(&r.Slippage, "slippage", 0.03*0.01, "")
	flagset.IntVar(&r.StartYear, "startyear", today.Year(), "")
	flagset.IntVar(&r.StartQuarter, "startquarter", 0, "")
	flagset.IntVar(&r.FinishYear, "finishyear", today.Year(), "")
	flagset.IntVar(&r.FinishQuarter, "finishquarter", 3, "")
	flagset.BoolVar(&r.MultiContract, "multy", true, "")
	flagset.Parse(args)

	var candleStorage = candlestorage.FromCandleInterval(cli.MapPath("~/TradingData"), r.TimeframeName, moex.Moscow)
	return historyreport.Show(candleStorage, r)
}
