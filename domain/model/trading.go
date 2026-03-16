package model

import "time"

type CandleFinishedEvent struct {
	Candle
}

type Signal struct {
	Name     string
	Deadline time.Time
	Price    float64
	Value    float64
}

type SignalEvent struct {
	Signal
}

type PlannedPosition struct {
	Security  Security
	Portfolio Portfolio
	Planned   int
}
