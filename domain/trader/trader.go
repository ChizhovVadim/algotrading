package trader

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

type Trader struct {
	logger     *slog.Logger
	inbox      chan any
	Broker     *MultyBroker
	signals    []*SignalService
	portfolios []*PortfolioService
	strategies []*StrategyService
}

func NewTrader(
	logger *slog.Logger,
) *Trader {
	return &Trader{
		logger: logger,
		inbox:  make(chan any),
		Broker: NewMultyBroker(),
	}
}

func (t *Trader) Close() error {
	return t.Broker.Close()
}

func (t *Trader) Inbox() chan<- any {
	return t.inbox
}

func (t *Trader) AddSignal(signal *SignalService) {
	t.signals = append(t.signals, signal)
}

func (t *Trader) AddStrategy(strategy *StrategyService) {
	t.strategies = append(t.strategies, strategy)
}

// Каждый сигнал торгуем в каждом портфеле
func (t *Trader) AddStrategiesForAllSignalPortfolioPairs() {
	for _, signal := range t.signals {
		for _, portfolio := range t.portfolios {
			t.AddStrategy(NewStrategyService(t.logger, portfolio.broker, portfolio.portfolio, signal.security, signal.name))
		}
	}
}

func (t *Trader) AddPortfolio(portfolio *PortfolioService) {
	t.portfolios = append(t.portfolios, portfolio)
}

func (t *Trader) checkStatus() {
	t.Broker.CheckStatus()

	for _, signal := range t.signals {
		signal.CheckStatus()
	}
	fmt.Println("Total signals:", len(t.signals))

	for _, portfolio := range t.portfolios {
		portfolio.CheckStatus()
	}
	fmt.Println("Total portfolios:", len(t.portfolios))

	for _, strategy := range t.strategies {
		strategy.CheckStatus()
	}
	fmt.Println("Total strategies:", len(t.strategies))
}

func (t *Trader) init(ctx context.Context) error {
	t.logger.Info("Strategies starting...")
	if err := t.Broker.Init(ctx); err != nil {
		return err
	}
	for _, portfolio := range t.portfolios {
		var err = portfolio.Init()
		if err != nil {
			return err
		}
	}
	for _, strategy := range t.strategies {
		var err = strategy.Init()
		if err != nil {
			return err
		}
	}
	// сигналы последние, тк они подписываются на бары
	for _, signal := range t.signals {
		var err = signal.Init()
		if err != nil {
			return err
		}
	}
	t.logger.Info("Strategies started.")
	return nil
}

func (t *Trader) Run(ctx context.Context) error {
	if err := t.init(ctx); err != nil {
		return err
	}
	return t.eventLoop(ctx)
}

func (t *Trader) onCandle(candle model.Candle) bool {
	var orderRegistered bool
	for _, signalStrategy := range t.signals {
		var signal = signalStrategy.OnCandle(candle)
		if signal.DateTime.IsZero() {
			continue
		}
		for _, strategy := range t.strategies {
			if strategy.OnSignal(signal) {
				orderRegistered = true
			}
		}
	}
	return orderRegistered
}

func (t *Trader) eventLoop(ctx context.Context) error {
	var shouldCheckStatus = time.After(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-shouldCheckStatus:
			shouldCheckStatus = nil
			t.checkStatus()
		case msg, ok := <-t.inbox:
			if !ok {
				t.inbox = nil
				continue
			}
			switch msg := msg.(type) {
			case model.ExitUserCmd:
				return nil
			case model.CheckStatusUserCmd:
				t.checkStatus()
			case model.Candle:
				if t.onCandle(msg) {
					if shouldCheckStatus == nil {
						shouldCheckStatus = time.After(10 * time.Second)
					}
				}
			}
		}
	}
}
