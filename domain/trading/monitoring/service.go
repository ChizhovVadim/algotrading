package monitoring

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

type INotifyService interface {
	Notify(message string) error
}

type Service struct {
	logger        *slog.Logger
	broker        model.IBroker
	notifyService INotifyService
}

func New(
	logger *slog.Logger,
	broker model.IBroker,
	notifyService INotifyService,
) *Service {
	return &Service{
		logger:        logger,
		broker:        broker,
		notifyService: notifyService,
	}
}

func (s *Service) Update(
	signals []model.Signal,
	positions []model.PlannedPosition,
) {
	var (
		warningCount int
		errorCount   int
	)

	var sb = &strings.Builder{}

	s.broker.WriteStatus(sb)

	for _, signal := range signals {
		fmt.Fprintf(sb, "%-16v %16v %8v %8.4f\n",
			signal.Name,
			signal.Deadline.Format("2006-01-02 15:04"),
			signal.Price,
			signal.Value,
		)
	}
	fmt.Fprintln(sb, "Total signals:", len(signals))

	var visitedPortfolios = make(map[string]struct{})
	for _, position := range positions {
		var portfolio = position.Portfolio
		if _, found := visitedPortfolios[portfolio.Portfolio]; found {
			continue
		}
		visitedPortfolios[portfolio.Portfolio] = struct{}{}

		var limits, err = s.broker.GetPortfolioLimits(portfolio)
		if err != nil {
			warningCount += 1
			fmt.Fprintf(sb, "%-10v %-10v %v\n",
				portfolio.Client,
				portfolio.Portfolio,
				err)
		} else {
			var varMargin = limits.AccVarMargin + limits.VarMargin
			fmt.Fprintf(sb, "%-10v %-10v start: %10.0f varmargin: %10.0f (%.1f) used: %.1f\n",
				portfolio.Client,
				portfolio.Portfolio,
				limits.StartLimitOpenPos,
				varMargin,
				varMargin/limits.StartLimitOpenPos*100,
				limits.UsedLimOpenPos/limits.StartLimitOpenPos*100,
			)
		}
	}
	fmt.Fprintln(sb, "Total portfolios:", len(visitedPortfolios))

	for _, position := range positions {
		var brokerPos, err = s.broker.GetPosition(position.Portfolio, position.Security)
		if err != nil {
			warningCount += 1
			fmt.Fprintf(sb, "%-10v %-10v %10v %v\n",
				position.Portfolio.Client,
				position.Portfolio.Portfolio,
				position.Security.Name,
				err)
		} else {
			var status string
			if position.Planned == int(brokerPos) {
				status = "+"
			} else {
				errorCount += 1
				status = "!"
			}
			fmt.Fprintf(sb, "%-10v %-10v %10v planned: %6v actual: %6v %v\n",
				position.Portfolio.Client,
				position.Portfolio.Portfolio,
				position.Security.Name,
				position.Planned,
				int(brokerPos),
				status)
		}
	}
	fmt.Fprintln(sb, "Total strategies:", len(positions))

	if warningCount != 0 || errorCount != 0 {
		fmt.Fprintf(sb, "Warnings: %v Errors: %v\n", warningCount, errorCount)
	} else {
	}

	var message = sb.String()
	fmt.Print(message)
	if errorCount != 0 && s.notifyService != nil {
		var err = s.notifyService.Notify(message)
		if err != nil {
			s.logger.Info("notifyService.Notify", "error", err)
		}
	}
}
