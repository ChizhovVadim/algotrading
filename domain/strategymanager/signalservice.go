package strategymanager

import (
	"context"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

type Signal struct {
	Name               string
	DateTime           time.Time
	SecurityCode       string
	Price              float64
	Prediction         float64
	ContractsPerAmount Optional[float64]
	Deadline           time.Time
}

type SizeConfig struct {
	LongLever  float64
	ShortLever float64
	MaxLever   float64
	Weight     float64
}

type IAdvisor interface {
	Add(dt time.Time, price float64) (prediction float64, ok bool)
}

type SignalService struct {
	logger         *slog.Logger
	name           string
	marketData     model.IMarketData
	security       model.Security
	candleInterval string
	advisor        IAdvisor
	sizeConfig     SizeConfig
	start          time.Time
	baseCandle     model.Candle
	lastSignal     Signal
}

func NewSignalService(
	logger *slog.Logger,
	name string,
	marketData model.IMarketData,
	security model.Security,
	candleInterval string,
	advisor IAdvisor,
	sizeConfig SizeConfig,
) *SignalService {
	logger = logger.With(
		"name", name,
		"security", security.Name)
	return &SignalService{
		logger:         logger,
		name:           name,
		marketData:     marketData,
		security:       security,
		candleInterval: candleInterval,
		advisor:        advisor,
		sizeConfig:     sizeConfig,
		start:          time.Now().Add(-10 * time.Minute),
	}
}

func (s *SignalService) Init() error {
	if err := s.AddHistoryCandles(s.marketData.GetLastCandles(s.security, s.candleInterval)); err != nil {
		return err
	}
	s.logger.Info("Init signal",
		"DateTime", s.lastSignal.DateTime,
		"Price", s.lastSignal.Price,
		"Prediction", s.lastSignal.Prediction,
	)
	return nil
}

func (s *SignalService) Subscribe() error {
	return s.marketData.SubscribeCandles(s.security, s.candleInterval)
}

func (s *SignalService) WriteStatus(w io.Writer) {
	fmt.Fprintf(w, "%-10v %10v %16v %8v %8.4f\n",
		s.name,
		s.security.Name,
		s.lastSignal.DateTime.Format("2006-01-02 15:04"),
		s.lastSignal.Price,
		s.lastSignal.Prediction,
	)
}

func (s *SignalService) OnCandle(candle model.Candle) Signal {
	// советник следит только за своими барами
	if !( /*TODO s.candleInterval == candle.Interval &&*/
	s.security.Code == candle.SecurityCode) {
		return Signal{}
	}
	var prediction, ok = s.advisor.Add(candle.DateTime, candle.ClosePrice)
	if !ok {
		return Signal{}
	}
	var freshCandle = candle.DateTime.After(s.start)
	if s.baseCandle.DateTime.IsZero() && freshCandle {
		s.baseCandle = candle
		s.logger.Debug("Init base price",
			"DateTime", s.baseCandle.DateTime,
			"Price", s.baseCandle.ClosePrice)
	}
	var prevPrediction = s.lastSignal.Prediction
	s.lastSignal = Signal{
		Name:         s.name,
		SecurityCode: s.security.Code,
		DateTime:     candle.DateTime,
		Price:        candle.ClosePrice,
		Prediction:   prediction,
	}
	if !s.baseCandle.DateTime.IsZero() {
		var position = applySize(prediction, s.sizeConfig)
		s.lastSignal.ContractsPerAmount.SetValue(position / (s.baseCandle.ClosePrice * s.security.Lever))
		s.lastSignal.Deadline = candle.DateTime.Add(9 * time.Minute) // от открытия бара или 4 минуты от закрытия.
	}
	if freshCandle {
		var level slog.Level
		if s.lastSignal.Prediction != prevPrediction {
			level = slog.LevelInfo
		} else {
			level = slog.LevelDebug
		}
		s.logger.Log(context.Background(), level, "New signal",
			"Signal", s.lastSignal)
	}
	return s.lastSignal
}

func (s *SignalService) AddHistoryCandles(historyCandles iter.Seq2[model.Candle, error]) error {
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
		s.lastSignal = Signal{
			Name:         s.name,
			SecurityCode: s.security.Code,
			DateTime:     candle.DateTime,
			Price:        candle.ClosePrice,
			Prediction:   prediction,
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

func applySize(pos float64, config SizeConfig) float64 {
	if pos > 0 {
		pos *= config.LongLever
	} else {
		pos *= config.ShortLever
	}
	pos = config.Weight * max(-config.MaxLever, min(config.MaxLever, pos))
	return pos
}
