package strategymanager

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

type StrategyManager struct {
	logger     *slog.Logger
	signals    []*SignalService
	portfolios []*PortfolioService
	strategies []*StrategyService
}

func NewStrategyManager(
	logger *slog.Logger,
) *StrategyManager {
	return &StrategyManager{
		logger: logger,
	}
}

func (m *StrategyManager) AddSignal(signal *SignalService) {
	m.signals = append(m.signals, signal)
}

func (m *StrategyManager) AddPortfolio(portfolio *PortfolioService) {
	m.portfolios = append(m.portfolios, portfolio)
}

func (m *StrategyManager) AddStrategy(strategy *StrategyService) {
	m.strategies = append(m.strategies, strategy)
}

// Каждый сигнал торгуем в каждом портфеле
func (m *StrategyManager) AddStrategiesForAllSignalPortfolioPairs() {
	for _, signal := range m.signals {
		for _, portfolio := range m.portfolios {
			m.AddStrategy(NewStrategyService(m.logger, portfolio.broker, portfolio.portfolio, signal.security, signal.name))
		}
	}
}

func (m *StrategyManager) Init(ctx context.Context) error {
	m.logger.Info("Strategies starting...")
	for _, signal := range m.signals {
		var err = signal.Init()
		if err != nil {
			return err
		}
	}
	for _, portfolio := range m.portfolios {
		var err = portfolio.Init()
		if err != nil {
			return err
		}
	}
	for _, strategy := range m.strategies {
		var err = strategy.Init()
		if err != nil {
			return err
		}
	}
	m.logger.Info("Strategies started.")
	return nil
}

// Подписываться нужно в отдельной горутине
func (m *StrategyManager) Subscribe(ctx context.Context) error {
	for _, signal := range m.signals {
		var err = signal.Subscribe()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *StrategyManager) OnCandle(candle model.Candle) bool {
	var orderRegistered bool
	for _, signalStrategy := range m.signals {
		var signal = signalStrategy.OnCandle(candle)
		if signal.DateTime.IsZero() {
			continue
		}
		for _, strategy := range m.strategies {
			if strategy.OnSignal(signal) {
				orderRegistered = true
			}
		}
	}
	return orderRegistered
}

func (m *StrategyManager) WriteStatus(w io.Writer) {
	for _, signal := range m.signals {
		signal.WriteStatus(w)
	}
	fmt.Fprintln(w, "Total signals:", len(m.signals))

	for _, portfolio := range m.portfolios {
		portfolio.WriteStatus(w)
	}
	fmt.Fprintln(w, "Total portfolios:", len(m.portfolios))

	for _, strategy := range m.strategies {
		strategy.WriteStatus(w)
	}
	fmt.Fprintln(w, "Total strategies:", len(m.strategies))
}
