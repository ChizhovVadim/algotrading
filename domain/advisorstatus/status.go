package advisorstatus

import (
	"fmt"
	"iter"

	"github.com/ChizhovVadim/algotrading/domain/advisor"
	"github.com/ChizhovVadim/algotrading/domain/model"
)

type ICandleStorage interface {
	Candles(securityCode string) iter.Seq2[model.Candle, error]
}

func ShowStatus(
	candleStorage ICandleStorage,
	advisorName string,
	securityName string,
	count int,
) error {
	var advisor = advisor.BuildTest(advisorName)
	var advices []float64 //collections.deque(maxlen=count)
	for candle, err := range candleStorage.Candles(securityName) {
		if err != nil {
			return err
		}
		var newPosition, ok = advisor.Add(candle.DateTime, candle.ClosePrice)
		if !ok {
			continue
		}
		if len(advices) == 0 || advices[len(advices)-1] != newPosition {
			advices = append(advices, newPosition)
		}
	}
	fmt.Println(advisorName, securityName)
	for _, item := range advices[max(0, len(advices)-count):] {
		fmt.Println(item)
	}
	return nil
}
