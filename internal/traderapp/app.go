package traderapp

import (
	"context"
	"log"
	"log/slog"

	"github.com/ChizhovVadim/algotrading/domain/trader"
	"github.com/ChizhovVadim/algotrading/internal/usercommands"
)

type TraderApp struct {
	logger    *slog.Logger
	apiLogger *log.Logger
	config    Config
	trader    *trader.Trader
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
		trader:    trader.NewTrader(logger),
	}
}

func (app *TraderApp) Close() error {
	return app.trader.Close()
}

func (app *TraderApp) Run() error {
	app.logger.Info("Application started.")
	defer app.logger.Info("Application closed.")

	if err := app.configure(); err != nil {
		return err
	}

	var ctx = context.Background()

	go func() {
		var err = usercommands.Read(ctx, app.trader.Inbox())
		if err != nil {
			app.logger.Error("usercommands.Read", "error", err)
			return
		}
	}()

	return app.trader.Run(ctx)
}
