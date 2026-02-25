package historyreport

import "fmt"

type timeRange struct {
	StartYear     int
	StartQuarter  int
	FinishYear    int
	FinishQuarter int
}

func quarterSecurityCodes(name string, tr timeRange) []string {
	var result []string
	for year := tr.StartYear; year <= tr.FinishYear; year++ {
		for quarter := 0; quarter < 4; quarter++ {
			if year == tr.StartYear && quarter < tr.StartQuarter {
				continue
			}
			if year == tr.FinishYear && quarter > tr.FinishQuarter {
				break
			}
			var securityCode = fmt.Sprintf("%v-%v.%02d", name, 3+quarter*3, year%100)
			result = append(result, securityCode)
		}
	}
	return result
}
