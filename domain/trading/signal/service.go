package signal

import (
	"context"
	"iter"
	"log/slog"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

type IAdvisor interface {
	Add(dt time.Time, price float64) (prediction float64, ok bool)
}

type Service struct {
	logger         *slog.Logger
	marketData     model.IMarketData
	name           string
	security       model.Security
	candleInterval string
	advisor        IAdvisor
	start          time.Time
	currentSignal  model.Signal
}

func New(
	logger *slog.Logger,
	marketData model.IMarketData,
	name string,
	security model.Security,
	candleInterval string,
	advisor IAdvisor,
) *Service {
	return &Service{
		logger:         logger.With("name", name),
		marketData:     marketData,
		name:           name,
		security:       security,
		candleInterval: candleInterval,
		advisor:        advisor,
		start:          time.Now(),
	}
}

func (s *Service) Current() model.Signal {
	return s.currentSignal
}

func (s *Service) Init() error {
	if err := s.AddHistoryCandles(s.marketData.GetLastCandles(s.security, s.candleInterval)); err != nil {
		return err
	}
	s.logger.Info("Init signal", "signal", s.currentSignal)
	return nil
}

func (s *Service) Subscribe() error {
	return s.marketData.SubscribeCandles(s.security, s.candleInterval)
}

func (s *Service) OnCandle(msg model.CandleFinishedEvent) model.SignalEvent {
	// советник следит только за своими барами
	if !( /*TODO s.candleInterval == candle.Interval &&*/
	s.security.Code == msg.SecurityCode) {
		return model.SignalEvent{}
	}
	var prediction, ok = s.advisor.Add(msg.DateTime, msg.ClosePrice)
	if !ok {
		return model.SignalEvent{}
	}

	var signalValueChanged = s.currentSignal.Value != prediction
	s.currentSignal = model.Signal{
		Name:     s.name,
		Deadline: msg.DateTime.Add(9 * time.Minute),
		Price:    msg.ClosePrice,
		Value:    prediction,
	}
	if s.currentSignal.Deadline.Before(s.start) {
		return model.SignalEvent{}
	}

	var level slog.Level
	if signalValueChanged {
		level = slog.LevelInfo
	} else {
		level = slog.LevelDebug
	}
	s.logger.Log(context.Background(), level, "New signal", "signal", s.currentSignal)

	return model.SignalEvent{Signal: s.currentSignal}
}

func (s *Service) AddHistoryCandles(historyCandles iter.Seq2[model.Candle, error]) error {
	var (
		firstCandle model.Candle
		lastCandle  model.Candle
		size        int
	)
	for candle, err := range historyCandles {
		if err != nil {
			return err
		}

		if size == 0 {
			firstCandle = candle
		}
		lastCandle = candle
		size += 1

		var prediction, ok = s.advisor.Add(candle.DateTime, candle.ClosePrice)
		if !ok {
			continue
		}
		s.currentSignal = model.Signal{
			Name:     s.name,
			Deadline: candle.DateTime.Add(9 * time.Minute), //9 минут от открытия бара, 4 минуты от закрытия бара.
			Price:    candle.ClosePrice,
			Value:    prediction,
		}
	}
	if size == 0 {
		s.logger.Warn("History candles empty")
	} else {
		s.logger.Debug("History candles",
			"First", firstCandle,
			"Last", lastCandle,
			"Size", size)
	}
	return nil
}
