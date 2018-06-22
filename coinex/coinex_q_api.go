package coinex

import (
	"errors"
	"strings"

	"github.com/blockcdn-go/exchange-sdk-go/utils"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

// GetAllSymbol 交易市场详细行情接口
func (c *Client) GetAllSymbol() ([]global.TradeSymbol, error) {
	data := []string{}
	r := plainRsp{Data: &data}

	err := c.httpReq("GET", "https://api.coinex.com/v1/market/list", nil, &r, false)
	if err != nil {
		return nil, err
	}
	if r.Code != 0 {
		return nil, errors.New(r.Message)
	}
	ret := []global.TradeSymbol{}
	for _, d := range data {
		base, quote := split(d)
		ret = append(ret, global.TradeSymbol{Base: base, Quote: quote})
	}
	return ret, nil
}

// GetDepth 获取深度行情
func (c *Client) GetDepth(req global.TradeSymbol) (global.Depth, error) {
	data := struct {
		Asks [][]interface{} `json:"asks"`
		Bids [][]interface{} `json:"bids"`
	}{}
	r := plainRsp{Data: &data}
	in := map[string]interface{}{}
	in["market"] = strings.ToUpper(req.Base + req.Quote)
	in["merge"] = 0
	err := c.httpReq("GET", "https://api.coinex.com/v1/market/depth", in, &r, false)
	if err != nil {
		return global.Depth{}, err
	}
	if r.Code != 0 {
		return global.Depth{}, errors.New(r.Message)
	}
	ret := global.Depth{Base: req.Base, Quote: req.Quote,
		Asks: []global.DepthPair{}, Bids: []global.DepthPair{}}
	for _, p := range data.Asks {
		if len(p) < 2 {
			continue
		}
		ret.Asks = append(ret.Asks, global.DepthPair{Price: utils.ToFloat(p[0]), Size: utils.ToFloat(p[1])})
	}
	for _, p := range data.Bids {
		if len(p) < 2 {
			continue
		}
		ret.Bids = append(ret.Bids, global.DepthPair{Price: utils.ToFloat(p[0]), Size: utils.ToFloat(p[1])})
	}
	return ret, nil
}

// GetKline 获取k线数据
func (c *Client) GetKline(req global.KlineReq) ([]global.Kline, error) {
	in := map[string]interface{}{}
	in["market"] = strings.ToUpper(req.Base + req.Quote)
	in["type"] = utils.Period2Suffix(req.Period, false)
	data := [][]interface{}{}
	r := plainRsp{Data: &data}
	err := c.httpReq("GET", "https://api.coinex.com/v1/market/kline", in, &r, false)
	if err != nil {
		return nil, err
	}
	if r.Code != 0 {
		return nil, errors.New(r.Message)
	}
	ret := []global.Kline{}
	for _, k := range data {
		if len(k) < 6 {
			continue
		}
		ret = append(ret, global.Kline{
			Base:      req.Base,
			Quote:     req.Quote,
			Timestamp: int64(utils.ToFloat(k[0])) * 1000,
			Open:      utils.ToFloat(k[1]),
			Close:     utils.ToFloat(k[2]),
			High:      utils.ToFloat(k[3]),
			Low:       utils.ToFloat(k[4]),
			Volume:    utils.ToFloat(k[5]),
		})
	}
	return ret, nil
}
