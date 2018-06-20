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

// GetAllSymbol 获取所有的可交易对
func (c *Client) GetAllSymbol() ([]global.TradeSymbol, error) {
	r := struct {
		Status string      `json:"status"`
		Data   []TradePair `json:"data"`
		Errmsg string      `json:"err-msg"`
	}{}
	e := c.doHTTP("GET", "/v1/common/symbols", nil, &r)
	if e != nil {
		return nil, e
	}
	if r.Status != "ok" {
		return nil, fmt.Errorf(r.Errmsg)
	}
	ir := []global.TradeSymbol{}
	for _, p := range r.Data {
		ir = append(ir, global.TradeSymbol{
			Base:  p.Base,
			Quote: p.Quote,
		})
	}
	return ir, nil
}

// GetDepth 获取深度行情
func (c *Client) GetDepth(sreq global.TradeSymbol) (global.Depth, error) {
	symbol := strings.ToLower(sreq.Base + sreq.Quote)
	in := map[string]string{}
	in["symbol"] = symbol
	in["type"] = "step0"

	r := struct {
		Status string `json:"status"`
		Data   struct {
			Asks [][]float64 `json:"asks"` //卖方深度
			Bids [][]float64 `json:"bids"` //买方深度
		} `json:"tick"`
		Errmsg string `json:"err-msg"`
	}{}
	err := c.doHTTP("GET", "/market/depth", in, &r)
	if err != nil {
		return global.Depth{}, err
	}
	if r.Status != "ok" {
		return global.Depth{}, errors.New(r.Errmsg)
	}
	ret := global.Depth{
		Base:  sreq.Base,
		Quote: sreq.Quote,
		Asks:  make([]global.DepthPair, 0, 5),
		Bids:  make([]global.DepthPair, 0, 5),
	}
	if len(r.Data.Asks) >= 2 && r.Data.Asks[0][0] > r.Data.Asks[1][0] {
		// 卖 倒序
		for end := len(r.Data.Asks); end != 0; end-- {
			if len(r.Data.Asks[end-1]) < 2 {
				continue
			}
			ret.Asks = append(ret.Asks, global.DepthPair{
				Price: r.Data.Asks[end-1][0],
				Size:  r.Data.Asks[end-1][1]})
		}
	} else {
		for i := 0; i < len(r.Data.Asks); i++ {
			if len(r.Data.Asks[i]) < 2 {
				continue
			}
			ret.Asks = append(ret.Asks, global.DepthPair{Price: r.Data.Asks[i][0],
				Size: r.Data.Asks[i][1]})
		}
	}

	// 买
	for i := 0; i < len(r.Data.Bids); i++ {
		if len(r.Data.Bids[i]) < 2 {
			continue
		}
		ret.Bids = append(ret.Bids, global.DepthPair{
			Price: r.Data.Bids[i][0],
			Size:  r.Data.Bids[i][1]})
	}
	return ret, nil
}

// GetKline websocket 查询kline
func (c *Client) GetKline(req global.KlineReq) ([]global.Kline, error) {
	conn, err := c.connect()
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

	c.mutex.Lock()
	err = conn.WriteJSON(kreq)
	c.mutex.Unlock()
	if err != nil {
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
