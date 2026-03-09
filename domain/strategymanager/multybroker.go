package strategymanager

import (
	"context"
	"fmt"
	"io"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

var _ model.IBroker = (*MultyBroker)(nil)

type MultyBroker struct {
	brokers map[string]model.IBroker
}

func NewMultyBroker() *MultyBroker {
	return &MultyBroker{
		brokers: make(map[string]model.IBroker),
	}
}

func (b *MultyBroker) Add(key string, broker model.IBroker) {
	b.brokers[key] = broker
}

func (b *MultyBroker) Get(key string) model.IBroker {
	return b.brokers[key]
}

func (b *MultyBroker) Init(ctx context.Context) error {
	for _, child := range b.brokers {
		var err = child.Init(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *MultyBroker) WriteStatus(w io.Writer) {
	// на golang случайный порядок в map
	for _, child := range b.brokers {
		child.WriteStatus(w)
	}
	fmt.Fprintln(w, "Total brokers:", len(b.brokers))
}

func (b *MultyBroker) GetPortfolioLimits(portfolio model.Portfolio) (model.PortfolioLimits, error) {
	return b.brokers[portfolio.Client].GetPortfolioLimits(portfolio)
}

func (b *MultyBroker) GetPosition(portfolio model.Portfolio, security model.Security) (float64, error) {
	return b.brokers[portfolio.Client].GetPosition(portfolio, security)
}

func (b *MultyBroker) RegisterOrder(order model.Order) error {
	return b.brokers[order.Portfolio.Client].RegisterOrder(order)
}

func (b *MultyBroker) Close() error {
	for _, broker := range b.brokers {
		broker.Close()
	}
	//TODO errors.Join
	return nil
}
