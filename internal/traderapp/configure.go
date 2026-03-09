package traderapp

import (
	"fmt"
	"log/slog"

	"github.com/ChizhovVadim/algotrading/domain/advisor"
	"github.com/ChizhovVadim/algotrading/domain/model"
	"github.com/ChizhovVadim/algotrading/domain/strategymanager"
	"github.com/ChizhovVadim/algotrading/internal/brokermock"
	"github.com/ChizhovVadim/algotrading/internal/brokerquik"
	"github.com/ChizhovVadim/algotrading/internal/candlestorage"
	"github.com/ChizhovVadim/algotrading/internal/cli"
	"github.com/ChizhovVadim/algotrading/internal/moex"
)

func (app *TraderApp) configure(inbox chan<- any) error {
	var activeClients = computeActiveClients(app.config)
	for _, brokerConfig := range app.config.Brokers {
		var clientKey = brokerConfig.Key
		if _, found := activeClients[clientKey]; !found {
			continue
		}
		switch brokerConfig.Type {
		case "mock":
			app.strategyManager.Broker.Add(clientKey, brokermock.New(app.logger, clientKey))
		case "quik":
			app.strategyManager.Broker.Add(clientKey, brokerquik.New(app.logger, app.apiLogger, clientKey, brokerConfig.Port, inbox))
		default:
			// кроме quik можно поддержать API finam/alor/T.
			return fmt.Errorf("broker type not supported %v", brokerConfig.Type)
		}
	}

	var marketData = app.strategyManager.Broker.Get(app.config.MarketData).(model.IMarketData)
	app.logger.Debug("MarketData initialized",
		"client", app.config.MarketData)

	var candleStorageFolder string
	if app.config.UseCandleStorage {
		candleStorageFolder = cli.MapPath("~/TradingData")
	}
	for _, signalConfig := range app.config.Signals {
		signal, err := configureSignal(app.logger, signalConfig, candleStorageFolder, marketData)
		if err != nil {
			return err
		}
		app.strategyManager.AddSignal(signal)
	}

	// Один и тот же портфель создаем один раз
	for _, portfolioConfig := range app.config.Portfolios {
		var portfolio = &strategymanager.Portfolio{
			Portfolio: model.Portfolio{
				Client:    portfolioConfig.Client,
				Firm:      portfolioConfig.Firm,
				Portfolio: portfolioConfig.Account,
			},
		}
		app.strategyManager.AddPortfolio(strategymanager.NewPortfolioService(app.logger, app.strategyManager.Broker, portfolio, portfolioConfig.MaxAmount, portfolioConfig.Weight))
	}

	// Каждый сигнал торгуем в каждом портфеле
	app.strategyManager.AddStrategiesForAllSignalPortfolioPairs()

	return nil
}

func configureSignal(
	logger *slog.Logger,
	signalConfig SignalConfig,
	candleStorageFolder string,
	marketData model.IMarketData,
) (*strategymanager.SignalService, error) {
	sec, err := moex.GetSecurityInfo(signalConfig.Security)
	if err != nil {
		return nil, err
	}
	var candleInterval = model.CandleIntervalMinutes5
	var advisor = advisor.BuildMain(logger, signalConfig.Advisor, signalConfig.StdVolatility)
	/*if candleStorageFolder != "" {
		var candleStorage = candlestorage.FromCandleInterval(candleStorageFolder, candleInterval, moex.Moscow)
		if err := advisor.AddCandles(candleStorage.Candles(sec.Name)); err != nil {
			return nil, err
		}
	}*/
	var signal = strategymanager.NewSignalService(logger, signalConfig.Advisor,
		marketData, sec, candleInterval, advisor, convertSizeConfig(signalConfig.SizeConfig))
	if candleStorageFolder != "" {
		var candleStorage = candlestorage.FromCandleInterval(candleStorageFolder, candleInterval, moex.Moscow)
		if err := signal.AddHistoryCandles(candleStorage.Candles(sec.Name)); err != nil {
			return nil, err
		}
	}
	return signal, nil
}

func convertSizeConfig(s SizeConfig) strategymanager.SizeConfig {
	return strategymanager.SizeConfig{
		LongLever:  s.LongLever,
		ShortLever: s.ShortLever,
		MaxLever:   s.MaxLever,
		Weight:     s.Weight,
	}
}

func computeActiveClients(config Config) map[string]bool {
	var activeClients = make(map[string]bool)
	for _, portfolioConfig := range config.Portfolios {
		activeClients[portfolioConfig.Client] = true
	}
	if config.MarketData != "" {
		activeClients[config.MarketData] = true
	}
	return activeClients
}
