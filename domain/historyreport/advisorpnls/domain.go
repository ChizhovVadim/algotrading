package advisorpnls

import (
	"iter"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

type IAdvisor interface {
	Add(dt time.Time, price float64) (prediction float64, ok bool)
}

type ICandleStorage interface {
	Candles(securityCode string) iter.Seq2[model.Candle, error]
}
