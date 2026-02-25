package candleprovider

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

type FinamProvider struct {
	getSecCode func(string) string
	periodCode string
	client     *http.Client
	loc        *time.Location
}

func NewFinam(
	getSecCode func(string) string,
	timeframe string,
	client *http.Client,
	loc *time.Location,
) (*FinamProvider, error) {
	var periodCode = finamTimeFrame(timeframe)
	if periodCode == "" {
		return nil, fmt.Errorf("finam timeFrameCode not found %v", timeframe)
	}
	return &FinamProvider{
		getSecCode: getSecCode,
		periodCode: periodCode,
		client:     client,
		loc:        loc,
	}, nil
}

func (srv *FinamProvider) Name() string {
	return "finam"
}

// TODO ctx
func (srv *FinamProvider) Load(securityName string, beginDate, endDate time.Time) ([]model.Candle, error) {
	var secCode = srv.getSecCode(securityName)
	if secCode == "" {
		return nil, fmt.Errorf("securityCode not found %v", securityName)
	}
	url, err := finamUrl(secCode, srv.periodCode, beginDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("url failed %w", err)
	}
	res, err := getCandlesMatastock(srv.client, url, srv.loc)
	if err != nil {
		return nil, fmt.Errorf("getCandlesMatastock %v %w", url, err)
	}
	return res, nil
}

func finamUrl(securityCode, periodCode string,
	beginDate, endDate time.Time) (string, error) {
	baseUrl, err := url.Parse("https://export.finam.ru/data.txt?d=d&market=14&f=data.txt&e=.txt&cn=data&dtf=1&tmf=1&MSOR=0&sep=1&sep2=1&datf=1&at=1")
	if err != nil {
		return "", err
	}

	params, err := url.ParseQuery(baseUrl.RawQuery)
	if err != nil {
		return "", err
	}

	params.Set("em", securityCode)
	params.Set("df", strconv.Itoa(beginDate.Day()))
	params.Set("mf", strconv.Itoa(int(beginDate.Month())-1))
	params.Set("yf", strconv.Itoa(beginDate.Year()))
	params.Set("dt", strconv.Itoa(endDate.Day()))
	params.Set("mt", strconv.Itoa(int(endDate.Month())-1))
	params.Set("yt", strconv.Itoa(endDate.Year()))
	params.Set("p", periodCode)

	baseUrl.RawQuery = params.Encode()
	return baseUrl.String(), nil
}

func finamTimeFrame(tf string) string {
	if tf == model.CandleIntervalMinutes5 {
		return "3"
	}
	if tf == model.CandleIntervalHourly {
		return "7"
	}
	if tf == model.CandleIntervalDaily {
		return "8"
	}
	return ""
}
