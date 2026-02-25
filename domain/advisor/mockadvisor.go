package advisor

import "time"

type MockAdvisor struct{}

func (adv *MockAdvisor) Add(dt time.Time, price float64) (float64, bool) {
	return 0, false
}
