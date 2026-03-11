package traderapp

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/model"
	"github.com/ChizhovVadim/algotrading/domain/strategymanager"
)

type TraderApp struct {
	logger          *slog.Logger
	apiLogger       *log.Logger
	config          Config
	broker          *MultyBroker
	strategyManager *strategymanager.StrategyManager
}

func New(
	logger *slog.Logger,
	apiLogger *log.Logger,
	config Config,
) *TraderApp {
	return &TraderApp{
		logger:          logger,
		apiLogger:       apiLogger,
		config:          config,
		broker:          NewMultyBroker(),
		strategyManager: strategymanager.NewStrategyManager(logger),
	}
}

func (app *TraderApp) Close() error {
	return app.broker.Close()
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
	if err := app.strategyManager.Init(ctx); err != nil {
		return err
	}
	app.checkStatus()

	go func() {
		var err = app.strategyManager.Subscribe(ctx)
		if err != nil {
			app.logger.Error("strategyManager.Subscribe", "error", err)
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
			case model.Candle:
				if app.strategyManager.OnCandle(msg) {
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

func (app *TraderApp) checkStatus() {
	app.logger.Debug("checkStatus started.")
	var sb = &strings.Builder{}
	app.broker.WriteStatus(sb)
	app.strategyManager.WriteStatus(sb)
	fmt.Print(sb.String()) //TODO SmtpWriter
	app.logger.Debug("checkStatus finished.")
}
