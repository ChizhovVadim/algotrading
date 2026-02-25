package equityreport

import (
	"math"
	"sort"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

type HprStatistcs struct {
	MonthHpr           float64
	StDev              float64
	AVaR               float64
	DayHprs            []model.DateSum
	MonthHprs          []model.DateSum
	YearHprs           []model.DateSum
	DrawdownInfo       DrawdownInfo
	ProfitableRating   []model.DateSum
	UnprofitableRating []model.DateSum
}

type DrawdownInfo struct {
	HighEquityDate      time.Time
	MaxDrawdown         float64
	LongestDrawdown     int
	CurrentDrawdown     float64
	CurrentDrawdownDays int
}

func ReportDailyResults(dailyResults []model.DateSum) {
	var stat = computeHprStatistcs(dailyResults)
	printHprReport(stat)
}

func computeHprStatistcs(hprs []model.DateSum) HprStatistcs {
	var report = HprStatistcs{}
	report.DayHprs = hprs
	report.MonthHprs = hprsByPeriod(hprs, firstDayOMonth)
	report.YearHprs = hprsByPeriod(hprs, firstDayOfYear)
	report.MonthHpr = math.Pow(totalHpr(hprs), 22.0/float64(len(hprs)))
	report.StDev = stDevHprs(hprs)
	report.DrawdownInfo = computeDrawdownInfo(hprs)

	var sortedHprs = make([]model.DateSum, len(hprs))
	copy(sortedHprs, hprs)
	sort.Slice(sortedHprs, func(i, j int) bool {
		return sortedHprs[i].Sum < sortedHprs[j].Sum
	})
	report.AVaR = meanBySum(sortedHprs[:len(sortedHprs)/20])
	report.ProfitableRating = sortedHprs[max(0, len(sortedHprs)-10):]
	report.UnprofitableRating = sortedHprs[:min(len(sortedHprs), 10)]

	return report
}

func hprsByPeriod(hprs []model.DateSum, period func(time.Time) time.Time) []model.DateSum {
	var result []model.DateSum
	for i, hpr := range hprs {
		if i == 0 || period(result[len(result)-1].Date) != period(hpr.Date) {
			result = append(result, hpr)
		} else {
			var item = &result[len(result)-1]
			item.Date = hpr.Date
			item.Sum *= hpr.Sum
		}
	}
	return result
}

func meanBySum(hprs []model.DateSum) float64 {
	var items = make([]float64, len(hprs))
	for i := range items {
		items[i] = hprs[i].Sum
	}
	mean, _ := moments(items)
	return mean
}

func totalHpr(source []model.DateSum) float64 {
	var result = 1.0
	for _, item := range source {
		result *= item.Sum
	}
	return result
}

func stDevHprs(source []model.DateSum) float64 {
	var x = make([]float64, len(source))
	for i := range source {
		x[i] = math.Log(source[i].Sum)
	}
	return stDev(x)
}

func computeDrawdownInfo(hprs []model.DateSum) DrawdownInfo {
	var currentSum = 0.0
	var maxSum = 0.0
	var longestDrawdown = 0
	var currentDrawdownDays = 0
	var maxDrawdown = 0.0
	var highEquityDate = hprs[0].Date

	for _, hpr := range hprs {
		currentSum += math.Log(hpr.Sum)
		if currentSum > maxSum {
			maxSum = currentSum
			highEquityDate = hpr.Date
		}
		if curDrawdownn := currentSum - maxSum; curDrawdownn < maxDrawdown {
			maxDrawdown = curDrawdownn
		}
		currentDrawdownDays = int(hpr.Date.Sub(highEquityDate) / (time.Hour * 24))
		if currentDrawdownDays > longestDrawdown {
			longestDrawdown = currentDrawdownDays
		}
	}

	return DrawdownInfo{
		HighEquityDate:      highEquityDate,
		LongestDrawdown:     longestDrawdown,
		CurrentDrawdownDays: currentDrawdownDays,
		MaxDrawdown:         math.Exp(maxDrawdown),
		CurrentDrawdown:     math.Exp(currentSum - maxSum),
	}
}
