package monitoring

import (
	"fmt"
	"io"
	"iter"
	"log/slog"
	"strings"
	"text/tabwriter"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

type IBrokerRegistry interface {
	All() iter.Seq2[string, model.IBroker]
	GetPortfolioLimits(portfolio model.Portfolio) (model.PortfolioLimits, error)
	GetPosition(portfolio model.Portfolio, security model.Security) (float64, error)
}

type INotifyService interface {
	Notify(message string) error
}

type Service struct {
	logger         *slog.Logger
	brokerRegistry IBrokerRegistry
	notifyService  INotifyService
}

func New(
	logger *slog.Logger,
	brokerRegistry IBrokerRegistry,
	notifyService INotifyService,
) *Service {
	return &Service{
		logger:         logger,
		brokerRegistry: brokerRegistry,
		notifyService:  notifyService,
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

	var totalBrokers int
	for client, broker := range s.brokerRegistry.All() {
		fmt.Fprintf(sb, "%-10s %T\n", client, broker)
		totalBrokers += 1
	}
	fmt.Fprintln(sb, "Total brokers:", totalBrokers)

	var w = newTabWriter(sb)
	fmt.Fprintf(w, "Signal\tDeadline\tPrice\tPosition\t\n")
	for _, signal := range signals {
		fmt.Fprintf(w, "%v\t%v\t%v\t%.4f\t\n",
			signal.Name,
			signal.Deadline.Format("2006-01-02 15:04"),
			signal.Price,
			signal.Value,
		)
	}
	w.Flush()
	fmt.Fprintln(sb, "Total signals:", len(signals))

	w = newTabWriter(sb)
	fmt.Fprintf(w, "Client\tPortfolio\tAmount\tVarMargin\tVarMarginRatio\tUsedRatio\t\n")
	var visitedPortfolios = make(map[string]struct{})
	for _, position := range positions {
		var portfolio = position.Portfolio
		if _, found := visitedPortfolios[portfolio.Portfolio]; found {
			continue
		}
		visitedPortfolios[portfolio.Portfolio] = struct{}{}

		var limits, err = s.brokerRegistry.GetPortfolioLimits(portfolio)
		if err != nil {
			//TODO
			continue
		}
		var varMargin = limits.AccVarMargin + limits.VarMargin
		fmt.Fprintf(w, "%v\t%v\t%.0f\t%.0f\t%.1f\t%.1f\t\n",
			portfolio.Client,
			portfolio.Portfolio,
			limits.StartLimitOpenPos,
			varMargin,
			varMargin/limits.StartLimitOpenPos*100,
			limits.UsedLimOpenPos/limits.StartLimitOpenPos*100,
		)
	}
	w.Flush()
	fmt.Fprintln(sb, "Total portfolios:", len(visitedPortfolios))

	w = newTabWriter(sb)
	fmt.Fprintf(w, "Client\tPortfolio\tSecurity\tPlanned\tActual\tStatus\t\n")
	for _, position := range positions {
		var brokerPos, err = s.brokerRegistry.GetPosition(position.Portfolio, position.Security)
		if err != nil {
			//TODO
			continue
		}
		var status string
		if position.Planned == int(brokerPos) {
			status = "+"
		} else {
			errorCount += 1
			status = "!"
		}
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\t\n",
			position.Portfolio.Client,
			position.Portfolio.Portfolio,
			position.Security.Name,
			position.Planned,
			int(brokerPos),
			status,
		)
	}
	w.Flush()
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

func newTabWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 1, ' ', tabwriter.AlignRight)
}
