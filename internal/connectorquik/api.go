package connectorquik

import (
	"fmt"
)

func (q *QuikConnector) IsConnected() (ResponseJson, error) {
	return q.MakeQuery("isConnected", "")
}

func (q *QuikConnector) MessageInfo(msg string) (ResponseJson, error) {
	return q.MakeQuery("message", msg)
}

func (q *QuikConnector) GetPortfolioInfoEx(
	firmId string,
	clientCode string,
	limitKind int,
) (ResponseJson, error) {
	return q.MakeQuery("getPortfolioInfoEx",
		fmt.Sprintf("%v|%v|%v", firmId, clientCode, limitKind))
}

func (q *QuikConnector) GetFuturesHolding(
	firmId string,
	accId string,
	secCode string,
	posType int,
) (ResponseJson, error) {
	return q.MakeQuery("getFuturesHolding",
		fmt.Sprintf("%v|%v|%v|%v", firmId, accId, secCode, posType))
}

func (q *QuikConnector) SendTransaction(req Transaction) (ResponseJson, error) {
	//Все значения должны передаваться в виде строк
	return q.MakeQuery("sendTransaction", req)
}

func (q *QuikConnector) GetLastCandles(
	classCode string,
	securityCode string,
	interval int,
	count int,
) ([]Candle, error) {
	var res []Candle
	var resp ResponseJson
	resp.Data = &res
	var err = q.Execute(RequestJson{
		Command: "get_candles_from_data_source",
		Data:    fmt.Sprintf("%v|%v|%v|%v", classCode, securityCode, interval, count),
	}, &resp)
	return res, err
}

func (q *QuikConnector) SubscribeCandles(
	classCode string,
	securityCode string,
	interval int,
) (ResponseJson, error) {
	return q.MakeQuery(
		"subscribe_to_candles",
		fmt.Sprintf("%v|%v|%v", classCode, securityCode, interval))
}

func (q *QuikConnector) UnsubscribeCandles(
	classCode string,
	securityCode string,
	interval int,
) (ResponseJson, error) {
	return q.MakeQuery(
		"unsubscribe_from_candles",
		fmt.Sprintf("%v|%v|%v", classCode, securityCode, interval))
}

func (q *QuikConnector) IsCandleSubscribed(
	classCode string,
	securityCode string,
	interval int,
) (ResponseJson, error) {
	return q.MakeQuery(
		"is_subscribed",
		fmt.Sprintf("%v|%v|%v", classCode, securityCode, interval))
}
