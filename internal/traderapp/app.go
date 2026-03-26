package traderapp

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/model"
	"github.com/ChizhovVadim/algotrading/domain/trading/brokermulty"
	"github.com/ChizhovVadim/algotrading/domain/trading/monitoring"
	"github.com/ChizhovVadim/algotrading/domain/trading/signal"
	"github.com/ChizhovVadim/algotrading/domain/trading/strategy"
)

type TraderApp struct {
	logger     *slog.Logger
	apiLogger  *log.Logger
	config     Config
	broker     *brokermulty.Service
	monitoring *monitoring.Service
	signals    []*signal.Service
	strategies []*strategy.Service
}

func New(
	logger *slog.Logger,
	apiLogger *log.Logger,
	config Config,
) *TraderApp {
	return &TraderApp{
		logger:    logger,
		apiLogger: apiLogger,
		config:    config,
	}
}

func (app *TraderApp) Close() error {
	if app.broker != nil {
		return app.broker.Close()
	}
	return nil
}

func (app *TraderApp) Run() error {
	app.logger.Info("Application started.")
	defer app.logger.Info("Application closed.")

	var ctx = context.Background()
	var marketData = make(chan any, 1)

	if err := app.configure(marketData); err != nil {
		return err
	}
	if err := app.broker.Init(ctx); err != nil {
		return err
	}
	for _, signal := range app.signals {
		var err = signal.Init()
		if err != nil {
			return err
		}
	}
	for _, strategy := range app.strategies {
		var err = strategy.Init()
		if err != nil {
			return err
		}
	}
	app.checkStatus()

	go func() {
		// подписываемся в отдельной горутине,
		// чтобы сразу начать читать бары и не заблокироваться.
		// IMarketData::SubscribeCandles должен быть потокобезопасным.
		var err = app.subscribe()
		if err != nil {
			app.logger.Error("app.subscribe", "error", err)
			return
		}
	}()

	var userCmds = make(chan any)
	go func() {
		var err = readUserCommands(ctx, userCmds)
		if err != nil {
			app.logger.Error("readUserCommands", "error", err)
			return
		}
	}()

	return app.eventLoop(ctx, marketData, userCmds)
}

func (app *TraderApp) eventLoop(
	ctx context.Context,
	marketData <-chan any,
	userCmds <-chan any,
) error {
	var shouldCheckStatus <-chan time.Time
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-shouldCheckStatus:
			shouldCheckStatus = nil
			app.checkStatus()
		case msg, ok := <-marketData:
			if !ok {
				marketData = nil
				continue
			}
			switch msg := msg.(type) {
			case model.CandleFinishedEvent:
				if app.OnCandle(msg) {
					if shouldCheckStatus == nil {
						shouldCheckStatus = time.After(10 * time.Second)
					}
				}
			}
		case msg, ok := <-userCmds:
			if !ok {
				userCmds = nil
				continue
			}
			switch msg.(type) {
			case ExitUserCmd:
				return nil
			case CheckStatusUserCmd:
				app.checkStatus()
			}
		}
	}
}

func (app *TraderApp) subscribe() error {
	for _, signal := range app.signals {
		var err = signal.Subscribe()
		if err != nil {
			return err
		}
	}
	return nil
}

func (app *TraderApp) OnCandle(candle model.CandleFinishedEvent) bool {
	var orderRegistered bool
	for _, signalStrategy := range app.signals {
		var signal = signalStrategy.OnCandle(candle)
		if signal.Deadline.IsZero() {
			continue
		}
		for _, strategy := range app.strategies {
			if strategy.OnSignal(signal) {
				orderRegistered = true
			}
		}
	}
	// TODO сделать StrategyRebalanceEvent для мониторинга?
	return orderRegistered
}

func (app *TraderApp) checkStatus() {
	var signals []model.Signal
	for _, signal := range app.signals {
		signals = append(signals, signal.Current())
	}
	var positions []model.PlannedPosition
	for _, strategy := range app.strategies {
		positions = append(positions, strategy.PlannedPosition())
	}
	app.monitoring.Update(signals, positions)
}
