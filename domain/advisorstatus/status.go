package advisorstatus

import (
	"fmt"
	"iter"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/advisor"
	"github.com/ChizhovVadim/algotrading/domain/algo"
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
	var history = algo.NewSlidingWindow[Signal](count)
	for candle, err := range candleStorage.Candles(securityName) {
		if err != nil {
			return err
		}
		var prediction, ok = advisor.Add(candle.DateTime, candle.ClosePrice)
		if !ok {
			continue
		}
		if history.Len() == 0 ||
			history.Item(history.Len()-1).Prediction != prediction {
			history.Add(Signal{
				DateTime:   candle.DateTime,
				Price:      candle.ClosePrice,
				Prediction: prediction,
			})
		}
	}
	fmt.Println(advisorName, securityName)
	for _, item := range history.Items() {
		fmt.Println(item)
	}
	return nil
}
