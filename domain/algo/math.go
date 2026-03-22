package algo

import "math"

func Mean(source []float64) float64 {
	var total float64
	for _, x := range source {
		total += x
	}
	return total / float64(len(source))
}

func Moments(source []float64) (mean, stDev float64) {
	var n = 0
	var M2 = 0.0

	for _, x := range source {
		n++
		var delta = x - mean
		mean += delta / float64(n)
		M2 += delta * (x - mean)
	}

	if n == 0 {
		return math.NaN(), math.NaN()
	}

	stDev = math.Sqrt(M2 / float64(n))
	return
}

// Алгоритм вычисления среднеквадратического отклонения за один проход
func StDev(source []float64) float64 {
	var _, stDev = Moments(source)
	return stDev
}
