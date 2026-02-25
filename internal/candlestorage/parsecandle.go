package candlestorage

import (
	"strconv"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

func parseCandleMetastock(record []string, loc *time.Location) (model.Candle, error) {
	d, err := time.ParseInLocation("20060102", record[2], loc)
	if err != nil {
		return model.Candle{}, err
	}
	t, err := strconv.Atoi(record[3])
	if err != nil {
		return model.Candle{}, err
	}
	var hour = t / 10000
	var min = (t / 100) % 100
	d = d.Add(time.Duration(hour)*time.Hour + time.Duration(min)*time.Minute)
	o, err := strconv.ParseFloat(record[4], 64)
	if err != nil {
		return model.Candle{}, err
	}
	h, err := strconv.ParseFloat(record[5], 64)
	if err != nil {
		return model.Candle{}, err
	}
	l, err := strconv.ParseFloat(record[6], 64)
	if err != nil {
		return model.Candle{}, err
	}
	c, err := strconv.ParseFloat(record[7], 64)
	if err != nil {
		return model.Candle{}, err
	}
	v, err := strconv.ParseFloat(record[8], 64)
	if err != nil {
		return model.Candle{}, err
	}
	return model.Candle{
		DateTime:   d,
		OpenPrice:  o,
		HighPrice:  h,
		LowPrice:   l,
		ClosePrice: c,
		Volume:     v}, nil
}
