package advisor

import "log/slog"

func BuildTest(name string) *MockAdvisor {
	return &MockAdvisor{}
}

func BuildMain(
	logger *slog.Logger,
	name string,
	stdVolatility float64,
) *MockAdvisor {
	return &MockAdvisor{}
}
