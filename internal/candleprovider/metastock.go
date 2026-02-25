package candleprovider

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

func getCandlesMatastock(client *http.Client, url string, loc *time.Location) ([]model.Candle, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getCandlesMatastock http status %v", resp.Status)
	}
	var result []model.Candle
	csv := csv.NewReader(resp.Body)
	csv.Read() // skip fst line
	for {
		rec, err := csv.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("getCandlesMatastock %w", err)
		}
		c, err := parseCandleMetastock(rec, loc)
		if err != nil {
			return nil, fmt.Errorf("getCandlesMatastock %v %w", rec, err)
		}
		result = append(result, c)
	}
	return result, nil
}

func parseCandleMetastock(record []string, loc *time.Location) (model.Candle, error) {
	d, err := time.ParseInLocation("20060102", record[2], loc)
	if err != nil {
		return model.Candle{}, err
	}
	t, err := strconv.Atoi(record[3])
	if err != nil {
		return model.Candle{}, err
	}
	var hour = t / 10000
	var min = (t / 100) % 100
	d = d.Add(time.Duration(hour)*time.Hour + time.Duration(min)*time.Minute)
	o, err := strconv.ParseFloat(record[4], 64)
	if err != nil {
		return model.Candle{}, err
	}
	h, err := strconv.ParseFloat(record[5], 64)
	if err != nil {
		return model.Candle{}, err
	}
	l, err := strconv.ParseFloat(record[6], 64)
	if err != nil {
		return model.Candle{}, err
	}
	c, err := strconv.ParseFloat(record[7], 64)
	if err != nil {
		return model.Candle{}, err
	}
	v, err := strconv.ParseFloat(record[8], 64)
	if err != nil {
		return model.Candle{}, err
	}
	return model.Candle{
		DateTime:   d,
		OpenPrice:  o,
		HighPrice:  h,
		LowPrice:   l,
		ClosePrice: c,
		Volume:     v}, nil
}
