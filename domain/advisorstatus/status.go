package advisorstatus

import (
	"fmt"
	"iter"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/advisor"
	"github.com/ChizhovVadim/algotrading/domain/model"
)

type ICandleStorage interface {
	Candles(securityCode string) iter.Seq2[model.Candle, error]
}

type Signal struct {
	DateTime   time.Time
	Price      float64
	Prediction float64
}

func ShowStatus(
	candleStorage ICandleStorage,
	advisorName string,
	securityName string,
	count int,
) error {
	var advisor = advisor.BuildTest(advisorName)
	var history []Signal //TODO collections.deque(maxlen=count)
	for candle, err := range candleStorage.Candles(securityName) {
		if err != nil {
			return err
		}
		var prediction, ok = advisor.Add(candle.DateTime, candle.ClosePrice)
		if !ok {
			continue
		}
		if len(history) == 0 || history[len(history)-1].Prediction != prediction {
			history = append(history, Signal{
				DateTime:   candle.DateTime,
				Price:      candle.ClosePrice,
				Prediction: prediction,
			})
		}
	}
	fmt.Println(advisorName, securityName)
	for _, item := range history[max(0, len(history)-count):] {
		fmt.Println(item)
	}
	return nil
}
