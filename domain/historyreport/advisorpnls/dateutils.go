package advisorpnls

import "time"

func fromOneDay(a, b time.Time) bool {
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func dateTimeToDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func IsAfterLongHolidays(l, r time.Time) bool {
	if fromOneDay(l, r) {
		return false
	}
	y, m, d := l.Date()
	if y == 2022 && m == time.February && d == 25 {
		// приостановка торгов из-за СВО. выйти заранее невозможно!
		return false
	}
	var startDate = dateTimeToDate(l).AddDate(0, 0, 1)
	var endDate = dateTimeToDate(r)
	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		var weekDay = d.Weekday()
		if weekDay != time.Sunday && weekDay != time.Saturday {
			//В промежутке между currentDate и nextDate был 1 не выходной (значит торговала Америка)
			return true
		}
	}
	return false
}
