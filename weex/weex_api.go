package weex

import (
	"errors"
	"strings"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

var deftSymbol = []global.TradeSymbol{
	{Base: "BTC", Quote: "USD"},
	{Base: "BCH", Quote: "USD"},
	{Base: "LTC", Quote: "USD"},
	{Base: "ETH", Quote: "USD"},
	{Base: "ZEC", Quote: "USD"},
	{Base: "DASH", Quote: "USD"},
	{Base: "ETC", Quote: "USD"},
	{Base: "BCH", Quote: "BTC"},
	{Base: "LTC", Quote: "BTC"},
	{Base: "ETH", Quote: "BTC"},
	{Base: "ZEC", Quote: "BTC"},
	{Base: "DASH", Quote: "BTC"},
	{Base: "ETC", Quote: "BTC"},
}

type weexRsp struct {
	Code int         `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
}

// GetAllSymbol 交易市场详细行情接口
func (c *Client) GetAllSymbol() ([]global.TradeSymbol, error) {
	d := []struct {
		Quote string `json:"buy_asset_type"`
		Base  string `json:"sell_asset_type"`
	}{}
	r := weexRsp{Data: &d}
	err := c.httpReq("GET", "https://www.weexpro.com/exchange/v1/market/info", nil, &r, false)
	if err != nil {
		return nil, err
	}
	if r.Code != 0 {
		return nil, errors.New(r.Msg)
	}
	ret := []global.TradeSymbol{}
	for _, s := range d {
		ret = append(ret, global.TradeSymbol{
			Base:  s.Base,
			Quote: s.Quote,
		})
	}
	return ret, nil
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
	sybmol := strings.ToLower(req.Base + req.Quote)
	in := make(map[string]interface{})
	in["market"] = sybmol
	in["type"] = period
	d := [][]interface{}{}
	r := weexRsp{Data: &d}
	err := c.httpReq("GET", "https://api.weex.com/v1/market/kline", in, &r, false)
	if err != nil {
		return nil, err
	}
	if r.Code != 0 {
		return nil, errors.New(r.Msg)
	}
	ret := []global.Kline{}
	for _, v := range d {
		if len(v) < 6 {
			continue
		}
		ret = append(ret, global.Kline{
			Base:      req.Base,
			Quote:     req.Quote,
			Timestamp: int64(toFloat(v[0])) * 1000,
			Open:      toFloat(v[1]),
			Close:     toFloat(v[2]),
			High:      toFloat(v[3]),
			Low:       toFloat(v[4]),
			Volume:    toFloat(v[5]),
		})
	}
	return ret, nil
}

// GetFund 获取帐号资金余额
func (c *Client) GetFund(global.FundReq) ([]global.Fund, error) {
	d := map[string]struct {
		Available string `json:"available"`
		Frozen    string `json:"frozen"`
	}{}
	r := weexRsp{Data: &d}
	err := c.httpReq("GET", "https://api.weex.com/v1/balance/", nil, &r, true)
	if err != nil {
		return nil, err
	}
	if r.Code != 0 {
		return nil, errors.New(r.Msg)
	}
	ret := []global.Fund{}
	for k, v := range d {
		ret = append(ret, global.Fund{
			Base:      k,
			Available: toFloat(v.Available),
			Frozen:    toFloat(v.Frozen),
		})
	}
	return ret, nil
}
