package strategymanager

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

type Portfolio struct {
	Portfolio       model.Portfolio
	AmountAvailable Optional[float64]
}

type PortfolioService struct {
	logger    *slog.Logger
	broker    model.IBroker
	portfolio *Portfolio
	maxAmount float64
	weight    float64
}

func NewPortfolioService(
	logger *slog.Logger,
	broker model.IBroker,
	portfolio *Portfolio,
	maxAmount float64,
	weight float64,
) *PortfolioService {
	logger = logger.With(
		"client", portfolio.Portfolio.Client,
		"portfolio", portfolio.Portfolio.Portfolio)
	return &PortfolioService{
		logger:    logger,
		broker:    broker,
		portfolio: portfolio,
		maxAmount: maxAmount,
		weight:    weight,
	}
}

func (s *PortfolioService) Init() error {
	var limits, err = s.broker.GetPortfolioLimits(s.portfolio.Portfolio)
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
	s.portfolio.AmountAvailable.SetValue(availableAmount)
	return nil
}

func (s *PortfolioService) WriteStatus(w io.Writer) {
	var limits, err = s.broker.GetPortfolioLimits(s.portfolio.Portfolio)
	if err != nil {
		fmt.Fprintf(w, "%-10v %-10v %v\n",
			s.portfolio.Portfolio.Client,
			s.portfolio.Portfolio.Portfolio,
			err)
		return
	}
	var varMargin = limits.AccVarMargin + limits.VarMargin
	var varMarginRatio = varMargin / limits.StartLimitOpenPos
	var usedRatio = limits.UsedLimOpenPos / limits.StartLimitOpenPos

	fmt.Fprintf(w, "%-10v %-10v start: %10.0f available: %10.0f varmargin: %10.0f (%.1f) used: %.1f\n",
		s.portfolio.Portfolio.Client,
		s.portfolio.Portfolio.Portfolio,
		limits.StartLimitOpenPos,
		s.portfolio.AmountAvailable.Value,
		varMargin,
		varMarginRatio*100,
		usedRatio*100,
	)
}
