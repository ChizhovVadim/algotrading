package brokerquik

import (
	"io"
	"log"

	"github.com/ChizhovVadim/algotrading/domain/model"
	"github.com/ChizhovVadim/algotrading/internal/connectorquik"
	"github.com/ChizhovVadim/algotrading/internal/moex"

	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"strconv"
	"sync/atomic"
)

var _ model.IBroker = (*QuikBroker)(nil)
var _ model.IMarketData = (*QuikBroker)(nil)

type QuikBroker struct {
	logger              *slog.Logger
	name                string
	connector           *connectorquik.QuikConnector
	marketDataCallbacks chan<- any
	nextTransactionId   int64
}

func New(
	logger *slog.Logger,
	apiLogger *log.Logger,
	name string,
	port int,
	marketDataCallbacks chan<- any,
) *QuikBroker {
	logger = logger.With(
		"client", name,
		"type", "quik")
	return &QuikBroker{
		logger:              logger,
		name:                name,
		connector:           connectorquik.New(apiLogger, port, 1),
		marketDataCallbacks: marketDataCallbacks,
		nextTransactionId:   calculateStartTransId(),
	}
}

func (b *QuikBroker) handleCallbacks(ctx context.Context, cj connectorquik.CallbackJson) {
	if cj.Command == "NewCandle" {
		if cj.Data != nil && b.marketDataCallbacks != nil {
			var newCandle connectorquik.Candle
			var err = json.Unmarshal(*cj.Data, &newCandle)
			if err != nil {
				return //err
			}
			// TODO можно фильтровать слишком ранние бары
			select {
			case <-ctx.Done():
				//return ctx.Err()
			case b.marketDataCallbacks <- convertToCandle(newCandle):
			}
		}
		return
	}
}

func (b *QuikBroker) Init(ctx context.Context) error {
	if err := b.connector.Init(ctx, b.handleCallbacks); err != nil {
		return err
	}
	resp, err := b.connector.IsConnected()
	if err != nil {
		return err
	}
	res, _ := connectorquik.ParseInt(resp.Data)
	if !(res == 1) {
		return errors.New("trader is not connected")
	}
	b.logger.Info("Init broker")
	return nil
}

func (b *QuikBroker) WriteStatus(w io.Writer) {
	fmt.Fprintf(w, "%-10s %-10s\n", b.name, "quik")
}

func (b *QuikBroker) Close() error {
	return b.connector.Close()
}

func (b *QuikBroker) GetPortfolioLimits(portfolio model.Portfolio) (model.PortfolioLimits, error) {
	resp, err := b.connector.GetPortfolioInfoEx(portfolio.Firm, portfolio.Portfolio, 0)
	if err != nil {
		return model.PortfolioLimits{}, err
	}
	var data = connectorquik.AsMap(resp.Data)
	if data == nil {
		return model.PortfolioLimits{}, errors.New("portfolio not found")
	}
	startLimitOpenPos, ok := connectorquik.ParseFloat(data["start_limit_open_pos"])
	if !ok {
		return model.PortfolioLimits{}, errors.New("parse start_limit_open_pos")
	}
	usedLimOpenPos, _ := connectorquik.ParseFloat(data["used_lim_open_pos"])
	varMargin, _ := connectorquik.ParseFloat(data["varmargin"])
	accVarMargin, _ := connectorquik.ParseFloat(data["fut_accured_int"])
	return model.PortfolioLimits{
		StartLimitOpenPos: startLimitOpenPos,
		UsedLimOpenPos:    usedLimOpenPos,
		VarMargin:         varMargin,
		AccVarMargin:      accVarMargin,
	}, nil
}

func (b *QuikBroker) GetPosition(portfolio model.Portfolio, security model.Security) (float64, error) {
	if security.ClassCode == moex.FuturesClassCode {
		resp, err := b.connector.GetFuturesHolding(portfolio.Firm, portfolio.Portfolio, security.Code, 0)
		if err != nil {
			return 0, err
		}
		var data = connectorquik.AsMap(resp.Data)
		if data == nil {
			b.logger.Debug("empty position",
				"client", portfolio.Client,
				"portfolio", portfolio.Portfolio,
				"security", security.Name,
			)
			return 0, nil
		}
		pos, ok := connectorquik.ParseFloat(data["totalnet"])
		if !ok {
			return 0, fmt.Errorf("GetFuturesHolding bad response")
		}
		return pos, nil
	} else {
		return 0, fmt.Errorf("not supported classcode %v", security.ClassCode)
	}
}

func (b *QuikBroker) RegisterOrder(order model.Order) error {
	var sPrice = formatPrice(order.Security.PriceStep, order.Security.PricePrecision, order.Price)
	b.logger.Info("RegisterOrder",
		"portfolio", order.Portfolio.Portfolio,
		"security", order.Security.Name,
		"volume", order.Volume,
		"price", sPrice)

	var transId = atomic.AddInt64(&b.nextTransactionId, 1)
	var strTransId = fmt.Sprintf("%v", transId)
	var trans = connectorquik.Transaction{
		TRANS_ID:    strTransId,
		ACTION:      "NEW_ORDER",
		SECCODE:     order.Security.Code,
		CLASSCODE:   order.Security.ClassCode,
		ACCOUNT:     order.Portfolio.Portfolio,
		PRICE:       sPrice,
		CLIENT_CODE: strTransId,
	}
	if order.Volume > 0 {
		trans.OPERATION = "B"
		trans.QUANTITY = strconv.Itoa(order.Volume)
	} else {
		trans.OPERATION = "S"
		trans.QUANTITY = strconv.Itoa(-order.Volume)
	}
	_, err := b.connector.SendTransaction(trans)
	return err
}

func (b *QuikBroker) GetLastCandles(security model.Security, timeframe string) iter.Seq2[model.Candle, error] {
	return func(yield func(model.Candle, error) bool) {
		var candles, err = b.getLastCandles_Impl(security, timeframe)
		if err != nil {
			yield(model.Candle{}, err)
			return
		}
		for _, item := range candles {
			var candle = convertToCandle(item)
			if !yield(candle, nil) {
				return
			}
		}
	}
}

func (b *QuikBroker) getLastCandles_Impl(security model.Security, timeframe string) ([]connectorquik.Candle, error) {
	var candleInterval, ok = quikTimeframe(timeframe)
	if !ok {
		return nil, fmt.Errorf("timeframe not supported %v", timeframe)
	}
	const count = 5_000 // Если не указывать размер, то может прийти слишком много баров и unmarshal большой json
	var candles, err = b.connector.GetLastCandles(security.ClassCode, security.Code, candleInterval, count)
	if err != nil {
		return nil, err
	}
	// последний бар за сегодня может быть не завершен
	if len(candles) > 0 &&
		isToday(candles[len(candles)-1].Datetime.ToTime(moex.Moscow)) {
		candles = candles[:len(candles)-1]
	}
	return candles, nil
}

func (b *QuikBroker) SubscribeCandles(security model.Security, timeframe string) error {
	var candleInterval, ok = quikTimeframe(timeframe)
	if !ok {
		return fmt.Errorf("timeframe not supported %v", timeframe)
	}
	// TODO Можно проверять вдруг уже подписаны.
	b.logger.Debug("SubscribeCandles",
		"security", security.Code,
		"timeframe", timeframe)
	_, err := b.connector.SubscribeCandles(security.ClassCode, security.Code, candleInterval)
	return err
}
