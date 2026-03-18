package brokermulty

import (
	"context"
	"iter"
	"maps"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

var _ model.IBroker = (*Service)(nil)

type Service struct {
	brokers map[string]model.IBroker
}

func New() *Service {
	return &Service{
		brokers: make(map[string]model.IBroker),
	}
}

func (b *Service) Add(key string, broker model.IBroker) {
	b.brokers[key] = broker
}

func (b *Service) Get(key string) model.IBroker {
	return b.brokers[key]
}

func (b *Service) All() iter.Seq2[string, model.IBroker] {
	return maps.All(b.brokers)
}

func (b *Service) Init(ctx context.Context) error {
	for _, child := range b.brokers {
		var err = child.Init(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Service) GetPortfolioLimits(portfolio model.Portfolio) (model.PortfolioLimits, error) {
	return b.brokers[portfolio.Client].GetPortfolioLimits(portfolio)
}

func (b *Service) GetPosition(portfolio model.Portfolio, security model.Security) (float64, error) {
	return b.brokers[portfolio.Client].GetPosition(portfolio, security)
}

func (b *Service) RegisterOrder(order model.Order) error {
	return b.brokers[order.Portfolio.Client].RegisterOrder(order)
}

func (b *Service) Close() error {
	for _, broker := range b.brokers {
		broker.Close()
	}
	//TODO errors.Join
	return nil
}
