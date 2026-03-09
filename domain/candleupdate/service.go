package candleupdate

import (
	"fmt"
	"iter"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

type ICandleStorage interface {
	Candles(securityCode string) iter.Seq2[model.Candle, error]
	Last(securityCode string) (model.Candle, error)
	Update(securityCode string, candles []model.Candle) error
}

type ICandleProvider interface {
	Load(securityName string, beginDate, endDate time.Time) ([]model.Candle, error)
}

type Service struct {
	logger         *slog.Logger
	candleStorage  ICandleStorage
	candleProvider ICandleProvider
	maxDays        int
}

func New(
	logger *slog.Logger,
	candleStorage ICandleStorage,
	candleProvider ICandleProvider,
	maxDays int,
) *Service {
	return &Service{
		logger:         logger,
		candleStorage:  candleStorage,
		candleProvider: candleProvider,
		maxDays:        maxDays,
	}
}

func (s *Service) UpdateGroup(securityNames []string) error {
	var secCodeFailed []string
	for _, securityName := range securityNames {
		var err = s.Update(securityName)
		if err != nil {
			secCodeFailed = append(secCodeFailed, securityName)
		}
		time.Sleep(1 * time.Second)
	}
	if len(secCodeFailed) != 0 {
		return fmt.Errorf("UpdateGroup failed size: %v securities: %v",
			len(secCodeFailed), secCodeFailed)
	}
	return nil
}

func (s *Service) Update(security string) error {
	var lastCandle, err = s.candleStorage.Last(security)
	if err != nil {
		return err
	}
	var beginDate, endDate = calcDates(s.maxDays, security, lastCandle.DateTime)
	candles, err := s.candleProvider.Load(security, beginDate, endDate)
	if err != nil {
		return err
	}
	if len(candles) == 0 {
		return fmt.Errorf("download empty %v", security)
	}
	//Последний бар за сегодня может быть еще не завершен
	if fromOneDay(time.Now(), candles[len(candles)-1].DateTime) {
		candles = candles[:len(candles)-1]
	}
	if !lastCandle.DateTime.IsZero() {
		candles = candlesAfter(candles, lastCandle.DateTime)
	}
	if len(candles) == 0 {
		s.logger.Info("No new candles",
			"security", security)
		return nil
	}
	s.logger.Info("New candles",
		"security", security,
		"size", len(candles),
		"first", candles[0],
		"last", candles[len(candles)-1])
	if !lastCandle.DateTime.IsZero() {
		// слишком большое изменение цены может быть ошибкой поставщика
		var err = checkPriceChange(lastCandle, candles[0])
		if err != nil {
			return err
		}
	}
	return s.candleStorage.Update(security, candles)
}

func candlesAfter(source []model.Candle, date time.Time) []model.Candle {
	for i, candle := range source {
		if candle.DateTime.After(date) {
			return source[i:]
		}
	}
	return nil
}

func calcDates(maxDays int, securityCode string, from time.Time) (beginDate, endDate time.Time) {
	if from.IsZero() {
		beginDate = calcStartDate(securityCode)
	} else {
		beginDate = from
	}
	var today = time.Now()
	endDate = today
	// ограничение на кол-во скачиваемых данных за раз
	if maxDays != 0 {
		var limitDate = beginDate.AddDate(0, 0, maxDays)
		if limitDate.Before(endDate) {
			endDate = limitDate
		}
	}
	return
}

func calcStartDate(securityCode string) time.Time {
	// Для квартального фьючерса качаем за 4 месяца до примерной экспирации
	return approxExpirationDate(securityCode).AddDate(0, -4, 0)
}

func checkPriceChange(x, y model.Candle) error {
	const Width = 0.25
	var closeChange = math.Abs(math.Log(x.ClosePrice / y.ClosePrice))
	var openChange = math.Abs(math.Log(x.ClosePrice / y.OpenPrice))
	if openChange >= Width && closeChange >= Width {
		return fmt.Errorf("big jump %v %v", x, y)
	}
	return nil
}

func approxExpirationDate(securityCode string) time.Time {
	// С 1 июля 2015, для новых серий по кот нет открытых позиций, все основные фьючерсы и опционы должны исполняться в 3-й четверг месяца
	// name-month.year
	var delim1 = strings.Index(securityCode, "-")
	if delim1 == -1 {
		return time.Time{}
	}
	var delim2 = strings.Index(securityCode, ".")
	if delim2 == -1 {
		return time.Time{}
	}
	month, err := strconv.Atoi(securityCode[delim1+1 : delim2])
	if err != nil {
		return time.Time{}
	}
	year, err := strconv.Atoi(securityCode[delim2+1:])
	if err != nil {
		return time.Time{}
	}
	var curYear = time.Now().Year()
	year = curYear - curYear%100 + year
	return time.Date(year, time.Month(month), 15, 0, 0, 0, 0, time.Local)
}

func fromOneDay(a, b time.Time) bool {
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
