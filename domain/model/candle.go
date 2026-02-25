package model

import "time"

const (
	CandleIntervalMinutes5 = "minutes5"
	CandleIntervalHourly   = "hourly"
	CandleIntervalDaily    = "daily"
)

type Candle struct {
	Interval     string
	SecurityCode string
	DateTime     time.Time
	OpenPrice    float64
	HighPrice    float64
	LowPrice     float64
	ClosePrice   float64
	Volume       float64
}
