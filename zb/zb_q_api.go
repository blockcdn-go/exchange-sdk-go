package zb

import (
	"errors"
	"strings"

	"github.com/blockcdn-go/exchange-sdk-go/global"
	"github.com/blockcdn-go/exchange-sdk-go/utils"
)

// GetAllSymbol 交易市场详细行情接口
func (c *Client) GetAllSymbol() ([]global.TradeSymbol, error) {
	r := map[string]interface{}{}
	err := c.httpReq("GET", "http://api.zb.com/data/v1/markets", nil, &r, false)
	ret := []global.TradeSymbol{}
	for k := range r {
		base, quote := split3(k)
		ret = append(ret, global.TradeSymbol{
			Base:  base,
			Quote: quote,
		})
	}
	return ret, err
}

// GetDepth 获取深度行情
func (c *Client) GetDepth(req global.TradeSymbol) (global.Depth, error) {

	arg := map[string]interface{}{}
	arg["market"] = strings.ToLower(req.Base + "_" + req.Quote)
	arg["size"] = 100

	r := struct {
		errInfo
		Asks [][]float64 `json:"asks"`
		Bids [][]float64 `json:"bids"`
	}{}
	err := c.httpReq("GET", "http://api.zb.com/data/v1/depth", arg, &r, false)
	if err != nil {
		return global.Depth{}, err
	}
	if r.errInfo.Code != 0 {
		return global.Depth{}, errors.New(r.errInfo.Message)
	}
	dp := global.Depth{
		Base:  req.Base,
		Quote: req.Quote,
		Asks:  []global.DepthPair{},
		Bids:  []global.DepthPair{},
	}
	for _, a := range r.Asks {
		if len(a) < 2 {
			continue
		}
		dp.Asks = append(dp.Asks, global.DepthPair{Price: a[0], Size: a[1]})
	}
	for _, b := range r.Bids {
		if len(b) < 2 {
			continue
		}
		dp.Bids = append(dp.Bids, global.DepthPair{Price: b[0], Size: b[1]})
	}
	return dp, nil
}

// GetKline 获取k线数据
func (c *Client) GetKline(req global.KlineReq) ([]global.Kline, error) {
	period := req.Period
	if strings.Contains(period, "m") {
		period = period + "in"
	} else if strings.Contains(period, "h") {
		period = period + "our"
	} else if strings.Contains(period, "d") {
		period = period + "ay"
	} else if strings.Contains(period, "w") {
		period = period + "eek"
	}
	arg := map[string]interface{}{}
	arg["market"] = strings.ToLower(req.Base + "_" + req.Quote)
	arg["type"] = period
	arg["size"] = utils.Ternary(req.Count == 0, 500, req.Count)
	r := struct {
		errInfo
		Data [][]float64 `json:"data"`
	}{}
	err := c.httpReq("GET", "http://api.zb.com/data/v1/kline", arg, &r, false)
	if err != nil {
		return nil, err
	}
	if r.errInfo.Code != 0 {
		return nil, errors.New(r.errInfo.Message)
	}
	kline := []global.Kline{}
	for _, k1 := range r.Data {
		if len(k1) < 6 {
			continue
		}
		kline = append(kline, global.Kline{
			Base:      req.Base,
			Quote:     req.Quote,
			Timestamp: int64(k1[0]),
			Open:      k1[1],
			High:      k1[2],
			Low:       k1[3],
			Close:     k1[4],
			Volume:    k1[5],
		})
	}
	return kline, nil
}
