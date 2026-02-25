package brokerquik

import (
	"math"
	"strconv"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/model"
	"github.com/ChizhovVadim/algotrading/internal/connectorquik"
	"github.com/ChizhovVadim/algotrading/internal/moex"
)

func calculateStartTransId() int64 {
	var hour, min, sec = time.Now().Clock()
	return 60*(60*int64(hour)+int64(min)) + int64(sec)
}

func formatPrice(priceStep float64, pricePrecision int, price float64) string {
	if priceStep != 0 {
		price = math.Round(price/priceStep) * priceStep
	}
	return strconv.FormatFloat(price, 'f', pricePrecision, 64)
}

func isToday(d time.Time) bool {
	var y1, m1, d1 = d.Date()
	var y2, m2, d2 = time.Now().Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func convertToCandle(item connectorquik.Candle) model.Candle {
	return model.Candle{
		Interval:     "TODO",
		SecurityCode: item.SecCode,
		DateTime:     item.Datetime.ToTime(moex.Moscow),
		OpenPrice:    item.Open,
		HighPrice:    item.High,
		LowPrice:     item.Low,
		ClosePrice:   item.Close,
		Volume:       item.Volume,
	}
}

func quikTimeframe(timeframe string) (int, bool) {
	if timeframe == model.CandleIntervalMinutes5 {
		return connectorquik.CandleIntervalM5, true
	}
	return 0, false
}
