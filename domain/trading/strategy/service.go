package strategy

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

type SizePolicy struct {
	LongLever  float64
	ShortLever float64
	MaxLever   float64
	Weight     float64
	MaxAmount  float64
}

type Service struct {
	logger          *slog.Logger
	broker          model.IBroker
	signalName      string
	security        model.Security
	portfolio       model.Portfolio
	sizePolicy      SizePolicy
	amountAvailable float64
	// планируемая позиция в портфеле (после исполнения заявки)
	plannedPosition int
	basePrice       model.Signal
}

func New(
	logger *slog.Logger,
	broker model.IBroker,
	signalName string,
	security model.Security,
	portfolio model.Portfolio,
	sizePolicy SizePolicy,
) *Service {
	return &Service{
		logger: logger.With(
			"client", portfolio.Client,
			"portfolio", portfolio.Portfolio,
			"security", security.Name,
			"signal", signalName),
		broker:     broker,
		signalName: signalName,
		security:   security,
		portfolio:  portfolio,
		sizePolicy: sizePolicy,
	}
}

func (s *Service) Init() error {
	limits, err := s.broker.GetPortfolioLimits(s.portfolio)
	if err != nil {
		return err
	}
	s.amountAvailable = limits.StartLimitOpenPos
	if s.sizePolicy.MaxAmount != 0 {
		s.amountAvailable = min(s.amountAvailable, s.sizePolicy.MaxAmount)
	}

	brokerPos, err := s.getBrokerPos()
	if err != nil {
		return err
	}
	s.plannedPosition = int(brokerPos)

	s.logger.Info("Init strategy",
		"Position", s.plannedPosition,
		"amount", limits.StartLimitOpenPos,
		"availableAmount", s.amountAvailable)
	return nil
}

func (s *Service) getBrokerPos() (float64, error) {
	return s.broker.GetPosition(s.portfolio, s.security)
}

func (s *Service) PlannedPosition() model.PlannedPosition {
	return model.PlannedPosition{
		Security:  s.security,
		Portfolio: s.portfolio,
		Planned:   s.plannedPosition,
	}
}

func (s *Service) OnSignal(msg model.SignalEvent) bool {
	var orderRegistered bool
	var err = s.on_signal_impl(msg, &orderRegistered)
	if err != nil {
		s.logger.Warn("OnSignal failed",
			"error", err)
	}
	return orderRegistered
}

func (s *Service) on_signal_impl(signal model.SignalEvent, orderRegistered *bool) error {
	// стратегия следит только за своими сигналами
	if !(signal.Name == s.signalName) {
		return nil
	}
	// считаем, что сигнал слишком старый
	if signal.Deadline.Before(time.Now()) {
		return nil
	}
	if s.basePrice.Deadline.IsZero() {
		s.basePrice = signal.Signal
		s.logger.Debug("Init base price", "signal", s.basePrice)
	}
	var idealPos = calcIdealPos(s.amountAvailable, signal.Value, s.sizePolicy, s.security, s.basePrice.Price)
	var volume = int(idealPos - float64(s.plannedPosition))
	// изменение позиции не требуется
	if volume == 0 {
		return nil
	}
	brokerPos, err := s.getBrokerPos()
	if err != nil {
		return err
	}
	if s.plannedPosition != int(brokerPos) {
		return fmt.Errorf("check position failed")
	}
	err = s.broker.RegisterOrder(model.Order{
		Portfolio: s.portfolio,
		Security:  s.security,
		Volume:    volume,
		Price:     priceWithSlippage(signal.Price, volume),
	})
	if err != nil {
		return err
	}
	s.plannedPosition += volume
	*orderRegistered = true
	return nil
}

func calcIdealPos(
	amount float64,
	prediction float64,
	sizePolicy SizePolicy,
	security model.Security,
	price float64,
) float64 {
	var pos = prediction
	if pos > 0 {
		pos *= sizePolicy.LongLever
	} else {
		pos *= sizePolicy.ShortLever
	}
	pos = max(-sizePolicy.MaxLever, min(sizePolicy.MaxLever, pos))
	return amount * sizePolicy.Weight * pos / (price * security.Lever)
}

func priceWithSlippage(price float64, volume int) float64 {
	const Slippage = 0.001
	if volume > 0 {
		return price * (1 + Slippage)
	} else {
		return price * (1 - Slippage)
	}
}
