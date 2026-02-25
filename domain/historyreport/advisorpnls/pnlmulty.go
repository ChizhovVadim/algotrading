package advisorpnls

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

func MultiContractHprs(
	indBuilder func() IAdvisor,
	candleStorage ICandleStorage,
	secCodes []string,
	slippage float64,
	skipPnl func(l, r time.Time) bool,
	concurrency int,
) ([]model.DateSum, error) {
	if len(secCodes) == 1 {
		return SingleContractHprs(indBuilder, candleStorage, secCodes[0], slippage, skipPnl)
	}

	var index int32 = -1
	var wg = &sync.WaitGroup{}
	var hprsByContracts = make([][]model.DateSum, len(secCodes))
	for threadIndex := 0; threadIndex < concurrency; threadIndex++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				var i = int(atomic.AddInt32(&index, 1))
				if i >= len(secCodes) {
					break
				}
				var securityCode = secCodes[i]
				var hprs, err = SingleContractHprs(
					indBuilder,
					candleStorage,
					securityCode,
					slippage,
					skipPnl)
				if err != nil {
					log.Println(err)
					continue
				}
				hprsByContracts[i] = hprs
			}
		}()
	}
	wg.Wait()

	return concatHprs(hprsByContracts), nil
}

func concatHprs(hprsByContracts [][]model.DateSum) []model.DateSum {
	var result []model.DateSum
	for _, hprs := range hprsByContracts {
		if len(hprs) == 0 {
			continue
		}

		if len(result) != 0 {
			// последний день предыдущего контракта может быть не полный
			result = result[:len(result)-1]
		}

		var last = time.Time{}
		if len(result) != 0 {
			last = result[len(result)-1].Date
		}
		for i := 0; i < len(hprs); i++ {
			if hprs[i].Date.After(last) {
				result = append(result, hprs[i:]...)
				break
			}
		}
	}
	return result
}
