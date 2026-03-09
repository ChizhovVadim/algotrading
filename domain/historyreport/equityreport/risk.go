package equityreport

import "github.com/ChizhovVadim/algotrading/domain/model"

func LimitStDev(stDev float64) func([]model.DateSum) bool {
	return func(source []model.DateSum) bool {
		return stDevHprs(source) <= stDev
	}
}

func OptimalLever(hprs []model.DateSum, riskSpecification func([]model.DateSum) bool) float64 {
	var minHpr = hprs[0].Sum
	for _, x := range hprs[1:] {
		if x.Sum < minHpr {
			minHpr = x.Sum
		}
	}
	var maxLever = 1.0 / (1.0 - minHpr)
	var bestHpr = 1.0
	var bestLever = 0.0
	const step = 0.1

	// Шибко умные могли бы использовать метод деления отрезка пополам
	for lever := step; lever <= maxLever; lever += step {
		var leverHprs = HprsWithLever(hprs, lever)
		if !riskSpecification(leverHprs) {
			break
		}
		var hpr = totalHpr(leverHprs)
		if hpr < bestHpr {
			break
		}
		bestHpr = hpr
		bestLever = lever
	}

	return bestLever
}

func HprsWithLever(source []model.DateSum, lever float64) []model.DateSum {
	var result = make([]model.DateSum, len(source))
	for i, item := range source {
		result[i] = model.DateSum{
			Date: item.Date,
			Sum:  1 + lever*(item.Sum-1),
		}
	}
	return result
}
