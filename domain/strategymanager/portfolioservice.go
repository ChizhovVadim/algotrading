package strategymanager

import (
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

type PortfolioService struct {
	logger          *slog.Logger
	broker          model.IBroker
	portfolio       model.Portfolio
	maxAmount       float64 //TODO? Optional
	weight          float64 //TODO? Optional
	amountAvailable Optional[float64]
}

func NewPortfolioService(
	logger *slog.Logger,
	broker model.IBroker,
	portfolio model.Portfolio,
	maxAmount float64,
	weight float64,
) *PortfolioService {
	logger = logger.With(
		"client", portfolio.Client,
		"portfolio", portfolio.Portfolio)
	return &PortfolioService{
		logger:    logger,
		broker:    broker,
		portfolio: portfolio,
		maxAmount: maxAmount,
		weight:    weight,
	}
}

func (s *PortfolioService) Init() error {
	var limits, err = s.broker.GetPortfolioLimits(s.portfolio)
	if err != nil {
		return err
	}
	var availableAmount = limits.StartLimitOpenPos
	if s.weight != 0 {
		availableAmount *= s.weight
	}
	if s.maxAmount != 0 {
		availableAmount = min(availableAmount, s.maxAmount)
	}
	s.logger.Info("Init portfolio",
		"amount", limits.StartLimitOpenPos,
		"availableAmount", availableAmount)
	s.amountAvailable.SetValue(availableAmount)
	return nil
}

func (s *PortfolioService) GetAmountAvailable() (float64, error) {
	if !s.amountAvailable.HasValue {
		return 0, errors.New("amountAvailable is none")
	}
	return s.amountAvailable.Value, nil
}

func (s *PortfolioService) WriteStatus(w io.Writer) {
	var limits, err = s.broker.GetPortfolioLimits(s.portfolio)
	if err != nil {
		fmt.Fprintf(w, "%-10v %-10v %v\n",
			s.portfolio.Client,
			s.portfolio.Portfolio,
			err)
		return
	}
	var varMargin = limits.AccVarMargin + limits.VarMargin
	var varMarginRatio = varMargin / limits.StartLimitOpenPos
	var usedRatio = limits.UsedLimOpenPos / limits.StartLimitOpenPos

	fmt.Fprintf(w, "%-10v %-10v start: %10.0f available: %10.0f varmargin: %10.0f (%.1f) used: %.1f\n",
		s.portfolio.Client,
		s.portfolio.Portfolio,
		limits.StartLimitOpenPos,
		s.amountAvailable.Value,
		varMargin,
		varMarginRatio*100,
		usedRatio*100,
	)
}
