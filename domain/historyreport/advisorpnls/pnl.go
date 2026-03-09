package advisorpnls

import (
	"math"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

// Вычисляет дневные доходности советника.
// Есть 2 способоа вычислять доходность за день.
// 1. Прибыль сбрасываем по окончании основной сессии. Тогда доходности будут похожи на отчет брокера, тк клиринг идет по окончании основной сессии.
// 2. Прибыль сбрасываем на открытии основной сессии. Будет больше похоже на правду, тк размер позиции пересчитываем на открытии основной сессии.
func SingleContractHprs(
	indBuilder func() IAdvisor,
	candleStorage ICandleStorage,
	secCode string,
	slippage float64,
	skipPnl func(l, r time.Time) bool,
) ([]model.DateSum, error) {
	var (
		result       []model.DateSum
		pnl          float64
		prevPosition float64
		baseCandle   model.Candle
		prevCandle   model.Candle
	)

	var ind = indBuilder()
	for candle, err := range candleStorage.Candles(secCode) {
		if err != nil {
			return nil, err
		}
		var newPosition, ok = ind.Add(candle.DateTime, candle.ClosePrice)
		if !ok {
			continue
		}

		if !prevCandle.DateTime.IsZero() {
			if !(skipPnl != nil && skipPnl(prevCandle.DateTime, candle.DateTime)) {
				pnl += prevPosition*(candle.ClosePrice-prevCandle.ClosePrice) -
					slippage*math.Abs(newPosition-prevPosition)*candle.ClosePrice
			}

			if !fromOneDay(prevCandle.DateTime, candle.DateTime) {
				if !baseCandle.DateTime.IsZero() {
					var hpr = 1.0 + pnl/baseCandle.ClosePrice
					result = append(result, model.DateSum{
						Date: dateTimeToDate(baseCandle.DateTime),
						Sum:  hpr,
					})
				}
				baseCandle = candle
				pnl = 0
			}
		}

		prevPosition = newPosition
		prevCandle = candle
	}

	if !baseCandle.DateTime.IsZero() {
		var hpr = 1.0 + pnl/baseCandle.ClosePrice
		result = append(result, model.DateSum{
			Date: dateTimeToDate(baseCandle.DateTime),
			Sum:  hpr,
		})
	}

	return result, nil
}
