package weex

import (
	"errors"
	"strings"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

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

// GetDepth 获取深度行情
func (c *Client) GetDepth(req global.TradeSymbol) (global.Depth, error) {
	sybmol := strings.ToLower(req.Base + req.Quote)
	in := make(map[string]interface{})
	in["market"] = sybmol
	in["merge"] = 0
	in["limit"] = 100
	d := struct {
		Asks [][]string `json:"asks"`
		Bids [][]string `json:"bids"`
	}{}
	r := weexRsp{Data: &d}
	err := c.httpReq("GET", "https://api.weex.com/v1/market/depth", in, &r, false)
	if err != nil {
		return global.Depth{}, err
	}
	if r.Code != 0 {
		return global.Depth{}, errors.New(r.Msg)
	}
	dr := global.Depth{
		Base:  req.Base,
		Quote: req.Quote,
		Asks:  []global.DepthPair{},
		Bids:  []global.DepthPair{},
	}
	for _, a := range d.Asks {
		if len(a) < 2 {
			continue
		}
		dr.Asks = append(dr.Asks, global.DepthPair{
			Price: toFloat(a[0]),
			Size:  toFloat(a[1]),
		})
	}
	for _, b := range d.Bids {
		if len(b) < 2 {
			continue
		}
		dr.Bids = append(dr.Bids, global.DepthPair{
			Price: toFloat(b[0]),
			Size:  toFloat(b[1]),
		})
	}
	return dr, nil
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
