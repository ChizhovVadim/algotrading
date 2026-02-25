package cli

import "time"

const FlagDateLayout = "2006-01-02"

// TODO А почему не type DateValue time.Time

type DateValue struct {
	Date time.Time // Лучше использовать встраивание?
}

func (d *DateValue) String() string {
	return d.Date.Format(FlagDateLayout)
}

func (d *DateValue) Set(s string) error {
	if s == "" {
		d.Date = time.Time{}
		return nil
	}
	date, err := time.Parse(FlagDateLayout, s)
	if err != nil {
		return err
	}
	d.Date = date
	return nil
}
