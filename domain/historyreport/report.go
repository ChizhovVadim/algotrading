package historyreport

import (
	"fmt"
	"iter"
	"runtime"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/advisor"
	"github.com/ChizhovVadim/algotrading/domain/historyreport/advisorpnls"
	"github.com/ChizhovVadim/algotrading/domain/historyreport/equityreport"
	"github.com/ChizhovVadim/algotrading/domain/model"
)

type ReportRequest struct {
	AdvisorName   string
	TimeframeName string
	SecurityName  string
	Lever         float64
	Slippage      float64
	StartYear     int
	StartQuarter  int
	FinishYear    int
	FinishQuarter int
	MultiContract bool
}

type ICandleStorage interface {
	Candles(securityCode string) iter.Seq2[model.Candle, error]
}

func Show(
	candleStorage ICandleStorage,
	r ReportRequest,
) error {
	var secCodes []string
	if r.MultiContract {
		var tr = timeRange{
			StartYear:     r.StartYear,
			StartQuarter:  r.StartQuarter,
			FinishYear:    r.FinishYear,
			FinishQuarter: r.FinishQuarter,
		}
		secCodes = quarterSecurityCodes(r.SecurityName, tr)
	} else {
		secCodes = []string{r.SecurityName}
	}

	var start = time.Now()
	defer func() {
		fmt.Println("Elapsed:", time.Since(start))
	}()

	var concurrency = runtime.GOMAXPROCS(0)
	hprs, err := advisorpnls.MultiContractHprs(func() advisorpnls.IAdvisor {
		return advisor.BuildTest(r.AdvisorName)
	}, candleStorage, secCodes, r.Slippage, advisorpnls.IsAfterLongHolidays, concurrency)
	if err != nil {
		return err
	}
	var lever = r.Lever
	if lever == 0 {
		lever = equityreport.OptimalLever(hprs, equityreport.LimitStDev(0.045))
	}
	hprs = equityreport.HprsWithLever(hprs, lever)
	fmt.Println("Отчет", r.AdvisorName, r.SecurityName)
	fmt.Printf("Плечо: %.1f\n", lever)
	equityreport.ReportDailyResults(hprs)
	return nil
}
