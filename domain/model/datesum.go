package model

import "time"

// Доходность торговой системы за один день
type DateSum struct {
	Date time.Time
	Sum  float64
}
