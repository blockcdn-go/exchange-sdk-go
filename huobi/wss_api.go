package huobi

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

// SubMarketKLine 查询市场K线图
// period 可选 1min, 5min, 15min, 30min, 60min, 1day, 1mon, 1week, 1year
func (c *WSSClient) SubMarketKLine(symbol string, period string) (<-chan []byte, error) {
	cid, conn, err := c.connect()
	if err != nil {
		return nil, err
	}

	topic := fmt.Sprintf("market.%s.kline.%s", symbol, period)
	req := struct {
		Topic string `json:"sub"`
		ID    string `json:"id"`
	}{topic, c.generateClientID()}

	err = conn.WriteJSON(req)
	if err != nil {
		c.Close()
		return nil, err
	}

	result := make(chan []byte)
	go c.start(topic, cid, result)
	return result, nil
}

// GetKline websocket 查询kline
func (c *WSSClient) GetKline(req global.KlineReq) ([]global.Kline, error) {
	_, conn, err := c.connect()
	if err != nil {
		return nil, err
	}
	period := req.Period
	if strings.Contains(period, "m") {
		period = period + "in"
	} else if period == "1h" {
		period = "60m"
	} else if strings.Contains(period, "d") {
		period = period + "ay"
	} else if strings.Contains(period, "w") {
		period = period + "eek"
	}
	symbol := strings.ToLower(req.Base + req.Quote)
	topic := fmt.Sprintf("market.%s.kline.%s", symbol, period)
	kreq := struct {
		Topic string `json:"req"`
		ID    string `json:"id"`
		From  int64  `json:"from,omitempty"`
		To    int64  `json:"to,omitempty"`
	}{Topic: topic, ID: c.generateClientID()}

	err = conn.WriteJSON(kreq)
	if err != nil {
		c.Close()
		return nil, err
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(msg)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	message, _ := ioutil.ReadAll(gz)
	rsp := struct {
		Status string  `json:"status"`
		Data   []Kline `json:"data"`
		Errmsg string  `json:"err-msg"`
	}{}
	err = json.Unmarshal(message, &rsp)
	if err != nil {
		return nil, err
	}
	if rsp.Status != "ok" {
		return nil, errors.New("huobipro websocket kline error:" + rsp.Errmsg)
	}
	ik := []global.Kline{}
	for _, k := range rsp.Data {
		ik = append(ik, global.Kline{
			Base:      k.Base,
			Quote:     k.Quote,
			Timestamp: int64(k.Timestamp),
			High:      k.High,
			Open:      k.Open,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
		})
	}
	return ik, nil
}

// SubMarketDepth 查询市场深度数据
// type 可选值：{ step0, step1, step2, step3, step4, step5 } （合并深度0-5）；
// step0时，不合并深度
func (c *WSSClient) SubMarketDepth(symbol string, typ string) (<-chan []byte, error) {
	cid, conn, err := c.connect()
	if err != nil {
		return nil, err
	}

	topic := fmt.Sprintf("market.%s.depth.%s", symbol, typ)
	req := struct {
		Topic string `json:"sub"`
		ID    string `json:"id"`
	}{topic, c.generateClientID()}

	err = conn.WriteJSON(req)
	if err != nil {
		c.Close()
		return nil, err
	}

	result := make(chan []byte)
	go c.start(topic, cid, result)
	return result, nil
}

// SubTradeDetail 查询交易详细数据
func (c *WSSClient) SubTradeDetail(symbol string) (<-chan []byte, error) {
	cid, conn, err := c.connect()
	if err != nil {
		return nil, err
	}

	topic := fmt.Sprintf("market.%s.trade.detail", symbol)
	req := struct {
		Topic string `json:"sub"`
		ID    string `json:"id"`
	}{topic, c.generateClientID()}

	err = conn.WriteJSON(req)
	if err != nil {
		c.Close()
		return nil, err
	}

	result := make(chan []byte)
	go c.start(topic, cid, result)
	return result, nil
}

// SubMarketDetail 查询市场详情数据
func (c *WSSClient) SubMarketDetail(symbol string) (<-chan []byte, error) {
	cid, conn, err := c.connect()
	if err != nil {
		return nil, err
	}

	topic := fmt.Sprintf("market.%s.detail", symbol)
	req := struct {
		Topic string `json:"sub"`
		ID    string `json:"id"`
	}{topic, c.generateClientID()}

	err = conn.WriteJSON(req)
	if err != nil {
		c.Close()
		return nil, err
	}

	result := make(chan []byte)
	go c.start(topic, cid, result)
	return result, nil
}
