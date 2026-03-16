package traderapp

import (
	"fmt"

	"github.com/ChizhovVadim/algotrading/domain/advisor"
	"github.com/ChizhovVadim/algotrading/domain/model"
	"github.com/ChizhovVadim/algotrading/domain/trading/brokermulty"
	"github.com/ChizhovVadim/algotrading/domain/trading/monitoring"
	"github.com/ChizhovVadim/algotrading/domain/trading/signal"
	"github.com/ChizhovVadim/algotrading/domain/trading/strategy"
	"github.com/ChizhovVadim/algotrading/internal/brokermock"
	"github.com/ChizhovVadim/algotrading/internal/brokerquik"
	"github.com/ChizhovVadim/algotrading/internal/candlestorage"
	"github.com/ChizhovVadim/algotrading/internal/cli"
	"github.com/ChizhovVadim/algotrading/internal/moex"
	"github.com/ChizhovVadim/algotrading/internal/notifyemail"
)

func (app *TraderApp) configure(marketDataCh chan<- any) error {
	app.broker = brokermulty.New()

	var notifyService monitoring.INotifyService
	if app.config.NotifyConfig.Enabled {
		var c = app.config.NotifyConfig
		notifyService = notifyemail.New(c.From, c.To, c.Password, c.Host, c.Port, "Trading")
	}
	app.monitoring = monitoring.New(app.logger, app.broker, notifyService)

	var activeClients = computeActiveClients(app.config)
	for _, brokerConfig := range app.config.Brokers {
		var clientKey = brokerConfig.Key
		if _, found := activeClients[clientKey]; !found {
			continue
		}
		switch brokerConfig.Type {
		case "mock":
			app.broker.Add(clientKey, brokermock.New(app.logger, clientKey))
		case "quik":
			app.broker.Add(clientKey, brokerquik.New(app.logger, app.apiLogger, clientKey, brokerConfig.Port, marketDataCh))
		default:
			// кроме quik можно поддержать API finam/alor/T.
			return fmt.Errorf("broker type not supported %v", brokerConfig.Type)
		}
	}

	var marketData = app.broker.Get(app.config.MarketData).(model.IMarketData)
	app.logger.Debug("MarketData initialized",
		"client", app.config.MarketData)

	for _, signalConfig := range app.config.Signals {
		var security, err = moex.GetSecurityInfo(signalConfig.Security)
		if err != nil {
			return err
		}
		var signalName = getSignalName(signalConfig)
		const candleInterval = model.CandleIntervalMinutes5
		var advisor = advisor.BuildMain(app.logger, signalConfig.Advisor, signalConfig.StdVolatility)
		var signal = signal.New(app.logger, marketData, signalName, security, candleInterval, advisor)
		if app.config.UseCandleStorage {
			var candleStorage = candlestorage.FromCandleInterval(cli.MapPath("~/TradingData"), candleInterval, moex.Moscow)
			if err := signal.AddHistoryCandles(candleStorage.Candles(security.Name)); err != nil {
				return err
			}
		}
		app.signals = append(app.signals, signal)

		// Каждый сигнал торгуем в каждом портфеле
		for _, portfolioConfig := range app.config.Portfolios {
			var portfolio = model.Portfolio{
				Client:    portfolioConfig.Client,
				Firm:      portfolioConfig.Firm,
				Portfolio: portfolioConfig.Account,
			}
			var strategy = strategy.New(app.logger, app.broker, signalName, security, portfolio,
				strategy.SizePolicy{
					LongLever:  signalConfig.SizeConfig.LongLever,
					ShortLever: signalConfig.SizeConfig.ShortLever,
					MaxLever:   signalConfig.SizeConfig.MaxLever,
					Weight:     signalConfig.SizeConfig.Weight,
					MaxAmount:  portfolioConfig.MaxAmount,
				})
			app.strategies = append(app.strategies, strategy)
		}
	}

	return nil
}

func getSignalName(signalConfig SignalConfig) string {
	return fmt.Sprintf("%v-%v", signalConfig.Advisor, signalConfig.Security)
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
