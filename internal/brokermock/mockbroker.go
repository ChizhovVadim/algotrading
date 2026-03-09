package brokermock

import (
	"context"
	"fmt"
	"io"
	"iter"
	"log/slog"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

var _ model.IBroker = (*MockBroker)(nil)
var _ model.IMarketData = (*MockBroker)(nil)

type MockBroker struct {
	logger    *slog.Logger
	name      string
	positions map[string]float64
}

func New(logger *slog.Logger, name string) *MockBroker {
	logger = logger.With(
		"client", name,
		"type", "mock")
	return &MockBroker{
		logger:    logger,
		name:      name,
		positions: make(map[string]float64),
	}
}

func (b *MockBroker) Init(context.Context) error {
	b.logger.Info("Init broker")
	return nil
}

func (b *MockBroker) WriteStatus(w io.Writer) {
	fmt.Fprintf(w, "%-10s %-10s\n", b.name, "mock")
}

func (b *MockBroker) GetPortfolioLimits(portfolio model.Portfolio) (model.PortfolioLimits, error) {
	return model.PortfolioLimits{
		StartLimitOpenPos: 1_000_000,
	}, nil
}

func (b *MockBroker) GetPosition(portfolio model.Portfolio, security model.Security) (float64, error) {
	return b.positions[b.positionKey(portfolio, security)], nil
}

func (b *MockBroker) RegisterOrder(order model.Order) error {
	b.logger.Info("RegisterOrder",
		"portfolio", order.Portfolio.Portfolio,
		"security", order.Security.Name,
		"volume", order.Volume,
		"price", order.Price)
	b.positions[b.positionKey(order.Portfolio, order.Security)] += float64(order.Volume)
	return nil
}

func (b *MockBroker) Close() error {
	return nil
}

func (b *MockBroker) positionKey(portfolio model.Portfolio, security model.Security) string {
	return portfolio.Portfolio + security.Code
}

func (b *MockBroker) GetLastCandles(security model.Security, timeframe string) iter.Seq2[model.Candle, error] {
	return func(yield func(model.Candle, error) bool) {}
}

func (b *MockBroker) SubscribeCandles(security model.Security, timeframe string) error {
	b.logger.Debug("SubscribeCandles",
		"security", security.Code,
		"timeframe", timeframe)
	return nil
}
