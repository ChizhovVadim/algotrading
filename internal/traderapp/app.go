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
		strategyManager: strategymanager.NewStrategyManager(logger),
	}
}

func (app *TraderApp) Close() error {
	return app.strategyManager.Close()
}

func (app *TraderApp) Run() error {
	app.logger.Info("Application started.")
	defer app.logger.Info("Application closed.")

	var ctx = context.Background()
	var inbox = make(chan any)

	if err := app.configure(inbox); err != nil {
		return err
	}

	if err := app.strategyManager.Init(ctx); err != nil {
		return err
	}

	go func() {
		var err = app.strategyManager.Subscribe(ctx)
		if err != nil {
			app.logger.Error("strategyManager.Subscribe", "error", err)
			return
		}
	}()

	go func() {
		var err = readUserCommands(ctx, inbox)
		if err != nil {
			app.logger.Error("readUserCommands", "error", err)
			return
		}
	}()

	return app.eventLoop(ctx, inbox)
}

func (app *TraderApp) eventLoop(ctx context.Context, inbox <-chan any) error {
	var shouldCheckStatus = time.After(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-shouldCheckStatus:
			shouldCheckStatus = nil
			app.checkStatus()
		case msg, ok := <-inbox:
			if !ok {
				inbox = nil
				continue
			}
			switch msg := msg.(type) {
			case ExitUserCmd:
				return nil
			case CheckStatusUserCmd:
				app.checkStatus()
			case model.Candle:
				if app.strategyManager.OnCandle(msg) {
					if shouldCheckStatus == nil {
						shouldCheckStatus = time.After(10 * time.Second)
					}
				}
			}
		}
	}
}

func (app *TraderApp) checkStatus() {
	var sb = &strings.Builder{}
	app.strategyManager.WriteStatus(sb)
	fmt.Print(sb.String()) //TODO SmtpWriter
}
