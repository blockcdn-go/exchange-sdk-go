package aicoin

import (
	"errors"
	"strings"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/utils"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

// AicoinGetDepth 获取深度行情
func (c *Client) AicoinGetDepth(ex string, req global.TradeSymbol) (global.Depth, error) {
	in := map[string]interface{}{}
	in["symbol"] = strings.ToLower(ex + req.Base + req.Quote)
	r := struct {
		Asks [][]interface{} `json:"asks"`
		Bids [][]interface{} `json:"bids"`
	}{}
	err := c.aicoinHTTPReq("GET", "https://www.aicoin.net.cn/api/second/depths", in, &r)
	if err != nil {
		return global.Depth{}, err
	}
	ret := global.Depth{Base: req.Base, Quote: req.Quote,
		Asks: []global.DepthPair{}, Bids: []global.DepthPair{}}
	for _, a := range r.Asks {
		if len(a) < 2 {
			continue
		}
		ret.Asks = append(ret.Asks, global.DepthPair{
			Price: utils.ToFloat(a[0]), Size: utils.ToFloat(a[1])})
	}
	for _, b := range r.Bids {
		if len(b) < 2 {
			continue
		}
		ret.Bids = append(ret.Bids, global.DepthPair{
			Price: utils.ToFloat(b[0]), Size: utils.ToFloat(b[1])})
	}
	return ret, nil
}

// AicoinGetKline 获取k线数据
func (c *Client) AicoinGetKline(ex string, req global.KlineReq) ([]global.Kline, error) {
	step := 0
	switch req.Period {
	case "1m":
		step = 60
		break
	case "5m":
		step = 60 * 5
		break
	case "15m":
		step = 60 * 15
		break
	case "30m":
		step = 60 * 30
		break
	case "1h":
		step = 60 * 60 * 1
		break
	case "12h":
		step = 60 * 60 * 12
		break
	case "1d":
		step = 60 * 60 * 24
		break
	case "1w":
		step = 60 * 60 * 24 * 7
		break
	default:
		return nil, errors.New("not support period")
	}
	s := ex + req.Base + req.Quote
	in := map[string]interface{}{}
	in["symbol"] = strings.ToLower(s)
	in["step"] = step
	r := struct {
		Data [][]interface{} `json:"data"`
	}{}
	err := c.aicoinHTTPReq("GET", "https://www.aicoin.net.cn/api/second/kline", in, &r)
	if err != nil {
		return nil, err
	}

	ret := []global.Kline{}

	for _, dd := range r.Data {
		if len(dd) < 6 {
			continue
		}
		ret = append(ret, global.Kline{
			Base:      req.Base,
			Quote:     req.Quote,
			Timestamp: int64(utils.ToFloat(dd[0])) * 1000,
			Open:      utils.ToFloat(dd[1]),
			Close:     utils.ToFloat(dd[4]),
			High:      utils.ToFloat(dd[2]),
			Low:       utils.ToFloat(dd[3]),
			Volume:    utils.ToFloat(dd[5]),
		})
	}

	return ret, nil
}

// AicoinGetLateTrade 获取最近成交信息
func (c *Client) AicoinGetLateTrade(ex string, req global.TradeSymbol) ([]global.LateTrade, error) {
	r, err := c.omnipotentTicker(ex, req)
	if err != nil {
		return nil, err
	}
	if _, ok := r["trades"]; !ok {
		return nil, errors.New("no filed trades")
	}

	lt := r["trades"].([]interface{})
	ret := []global.LateTrade{}
	for _, l := range lt {
		ll := l.(map[string]interface{})
		dt := time.Unix(int64(int64(utils.ToFloat(ll["date"]))), 0).Format("2006-01-02 03:04:05 PM")
		ret = append(ret, global.LateTrade{
			Base:      req.Base,
			Quote:     req.Quote,
			DateTime:  dt,
			Num:       utils.ToFloat(ll["amount"]),
			Price:     utils.ToFloat(ll["price"]),
			Dircetion: utils.Ternary(utils.ToString(ll["trade_type"]) == "ask", "sell", "buy").(string),
		})
		ret[len(ret)-1].Total = ret[len(ret)-1].Price * ret[len(ret)-1].Num
	}
	return ret, nil
}

// AicoinGetTicker 获取ticker数据
func (c *Client) AicoinGetTicker(ex string, req global.TradeSymbol) (global.Ticker, error) {
	r, err := c.omnipotentTicker(ex, req)
	if err != nil {
		return global.Ticker{}, err
	}
	if _, ok := r["ticker"]; !ok {
		return global.Ticker{}, errors.New("no filed ticker")
	}

	tk := r["ticker"].(map[string]interface{})
	return global.Ticker{
		Base:               req.Base,
		Quote:              req.Quote,
		PriceChange:        utils.ToFloat(tk["diff"]),
		PriceChangePercent: utils.ToFloat(tk["degree24h"]),
		LastPrice:          utils.ToFloat(tk["last"]),
		HighPrice:          utils.ToFloat(tk["hight"]),
		LowPrice:           utils.ToFloat(tk["low"]),
		Volume:             utils.ToFloat(tk["vol24h"]),
	}, nil
}

//
func (c *Client) omnipotentTicker(ex string, req global.TradeSymbol) (map[string]interface{}, error) {
	in := map[string]interface{}{}
	in["symbol"] = strings.ToLower(ex + req.Base + req.Quote)
	r := map[string]interface{}{}
	err := c.aicoinHTTPReq("GET", "https://www.aicoin.net.cn/api/second/tickers", in, &r)
	return r, err
}
