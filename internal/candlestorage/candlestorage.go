package candlestorage

import (
	"encoding/csv"
	"io"
	"iter"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ChizhovVadim/algotrading/domain/model"
)

type CandleStorage struct {
	folderPath string
	loc        *time.Location
}

func NewCandleStorageByPath(
	folderPath string,
	loc *time.Location,
) *CandleStorage {
	return &CandleStorage{
		folderPath: folderPath, //os.MkdirAll(folderPath, os.ModePerm)
		loc:        loc,
	}
}

func FromCandleInterval(
	folderPath string,
	candleInterval string,
	loc *time.Location,
) *CandleStorage {
	return NewCandleStorageByPath(filepath.Join(folderPath, candleInterval), loc)
}

func (srv *CandleStorage) fileName(securityCode string) string {
	return filepath.Join(srv.folderPath, securityCode+".txt")
}

func (srv *CandleStorage) Candles(
	securityCode string,
) iter.Seq2[model.Candle, error] {
	return func(yield func(model.Candle, error) bool) {
		var path = srv.fileName(securityCode)
		var file, err = os.Open(path)
		if err != nil {
			yield(model.Candle{}, err)
			return
		}
		defer file.Close()
		var reader = csv.NewReader(file)
		//reader.Comma = ';'
		reader.Read()
		for {
			rec, err := reader.Read()
			if err != nil {
				if err == io.EOF {
					return
				}
				yield(model.Candle{}, err)
				return
			}
			candle, err := parseCandleMetastock(rec, srv.loc)
			if err != nil {
				yield(model.Candle{}, err)
				return
			}
			if !yield(candle, nil) {
				return
			}
		}
	}
}

func (srv *CandleStorage) Last(securityCode string) (model.Candle, error) {
	path := srv.fileName(securityCode)
	exists, err := isPathExists(path)
	if err != nil {
		return model.Candle{}, err
	}
	if !exists {
		return model.Candle{}, nil
	}

	var result model.Candle
	for c, err := range srv.Candles(securityCode) {
		if err != nil {
			return model.Candle{}, err
		}
		result = c
	}
	return result, nil
}

// Дописывает в конец файла
func (srv *CandleStorage) Update(securityCode string, candles []model.Candle) error {
	f, err := os.OpenFile(srv.fileName(securityCode), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	csv := csv.NewWriter(f)
	//TODO header
	for _, c := range candles {
		record := []string{
			securityCode,
			"5",
			c.DateTime.Format("20060102"),
			strconv.Itoa(100 * (100*c.DateTime.Hour() + c.DateTime.Minute())),
			strconv.FormatFloat(c.OpenPrice, 'f', -1, 64),
			strconv.FormatFloat(c.HighPrice, 'f', -1, 64),
			strconv.FormatFloat(c.LowPrice, 'f', -1, 64),
			strconv.FormatFloat(c.ClosePrice, 'f', -1, 64),
			strconv.FormatFloat(c.Volume, 'f', -1, 64),
		}
		err := csv.Write(record)
		if err != nil {
			return err
		}
	}
	csv.Flush()
	return csv.Error()
}

func isPathExists(path string) (bool, error) {
	_, err := os.Lstat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	// error other than not existing e.g. permission denied
	return false, err
}
