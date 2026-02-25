package equityreport

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

const dateFormatLayout = "02.01.2006"

func printHprReport(report HprStatistcs) {
	var w = newTabWriter()
	fmt.Fprintf(w, "Ежемесячная доходность\t%.1f%%\t\n", (report.MonthHpr-1)*100)
	fmt.Fprintf(w, "Среднеквадратичное отклонение доходности за день\t%.1f%%\t\n", report.StDev*100)
	fmt.Fprintf(w, "Средний убыток в день среди 5%% худших дней\t%.1f%%\t\n", (report.AVaR-1)*100)
	printDrawdownInfo(w, report.DrawdownInfo)
	w.Flush()

	fmt.Println("Доходности по дням")
	printHprs(report.DayHprs[max(0, len(report.DayHprs)-20):])

	fmt.Println("Доходности по месяцам")
	printHprs(report.MonthHprs)

	fmt.Println("Доходности по годам")
	printHprs(report.YearHprs)

	/*fmt.Println("Самые прибыльные дни")
	printHprs(report.ProfitableRating)

	fmt.Println("Самые убыточные дни")
	printHprs(report.UnprofitableRating)*/
}

func newTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)
}

func printHprs(source []model.DateSum) {
	var w = newTabWriter()
	for _, item := range source {
		fmt.Fprintf(w, "%v\t%.1f%%\t\n", item.Date.Format(dateFormatLayout), (item.Sum-1)*100)
	}
	w.Flush()
}

func printDrawdownInfo(w io.Writer, info DrawdownInfo) {
	fmt.Fprintf(w, "Максимальная просадка\t%.1f%%\t\n", hprPercent(info.MaxDrawdown))
	fmt.Fprintf(w, "Продолжительная просадка\t%v дн.\t\n", info.LongestDrawdown)
	fmt.Fprintf(w, "Текущая просадка\t%.1f%% %v дн.\t\n", hprPercent(info.CurrentDrawdown), info.CurrentDrawdownDays)
	fmt.Fprintf(w, "Дата максимума эквити\t%v\t\n", info.HighEquityDate.Format(dateFormatLayout))
}

func hprPercent(hpr float64) float64 {
	return (hpr - 1) * 100
}
