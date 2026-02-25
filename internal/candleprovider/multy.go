package candleprovider

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/model"
	"github.com/ChizhovVadim/algotrading/internal/moex"
)

type SecurityCode struct {
	Code      string `xml:",attr"`
	FinamCode string `xml:",attr"`
	MfdCode   string `xml:",attr"`
}

type ICandleProvider interface {
	Name() string
	Load(securityName string, beginDate, endDate time.Time) ([]model.Candle, error)
}

type MultyProvider struct {
	secCodes        map[string]SecurityCode
	candleInterval  string
	candleProviders []ICandleProvider
}

func NewMultyProvider(
	secCodes []SecurityCode,
	candleInterval string,
) *MultyProvider {
	return &MultyProvider{
		secCodes:       toMap(secCodes),
		candleInterval: candleInterval,
	}
}

func (p *MultyProvider) AddSource(providerName string) error {
	if providerName == "finam" {
		var finam, err = NewFinam(
			func(secCode string) string { return p.secCodes[secCode].FinamCode },
			p.candleInterval, &http.Client{Timeout: 25 * time.Second}, moex.Moscow)
		if err != nil {
			return err
		}
		p.candleProviders = append(p.candleProviders, finam)
		return nil
	}
	if providerName == "mfd" {
		var mfd, err = NewMfd(
			func(secCode string) string { return p.secCodes[secCode].MfdCode },
			p.candleInterval, &http.Client{Timeout: 25 * time.Second}, moex.Moscow)
		if err != nil {
			return err
		}
		p.candleProviders = append(p.candleProviders, mfd)
		return nil
	}
	return fmt.Errorf("bad candle provider %v", providerName)
}

func (p *MultyProvider) Load(securityName string, beginDate, endDate time.Time) ([]model.Candle, error) {
	var errs []error
	for _, provider := range p.candleProviders {
		var res, err = provider.Load(securityName, beginDate, endDate)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		return res, nil
	}
	return nil, errors.Join(errs...)
}

func toMap(secCodes []SecurityCode) map[string]SecurityCode {
	var res = make(map[string]SecurityCode)
	for _, secCode := range secCodes {
		res[secCode.Code] = secCode
	}
	return res
}
