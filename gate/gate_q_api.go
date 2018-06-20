package gate

import (
	"fmt"
	"strings"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

// GetAllSymbol 交易市场详细行情接口
func (c *Client) GetAllSymbol() ([]global.TradeSymbol, error) {
	var result struct {
		Result string               `json:"result"`
		Data   []MarketListResponse `json:"data"`
	}
	e := c.httpReq("GET", "/api2/1/marketlist", nil, &result)
	if e != nil {
		return nil, e
	}
	r := []global.TradeSymbol{}
	for _, s := range result.Data {
		r = append(r, global.TradeSymbol{
			Base:  s.CurrA,
			Quote: s.CurrB,
		})
	}
	return r, nil
}

// GetDepth 获取深度行情
func (c *Client) GetDepth(sreq global.TradeSymbol) (global.Depth, error) {
	symbol := strings.ToLower(sreq.Base + "_" + sreq.Quote)
	path := fmt.Sprintf("/api2/1/orderBook/%s", symbol)
	t := struct {
		Asks [][]float64 `json:"asks"` //卖方深度
		Bids [][]float64 `json:"bids"` //买方深度
	}{}

	e := c.httpReq("GET", path, nil, &t)
	if e != nil {
		return global.Depth{}, e
	}

	r := global.Depth{
		Base:  sreq.Base,
		Quote: sreq.Quote,
		Asks:  []global.DepthPair{},
		Bids:  []global.DepthPair{},
	}

	if len(t.Asks) >= 2 && t.Asks[0][0] > t.Asks[1][0] {
		// 卖 倒序
		for end := len(t.Asks); end != 0; end-- {
			if len(t.Asks[end-1]) < 2 {
				continue
			}
			r.Asks = append(r.Asks, global.DepthPair{
				Price: t.Asks[end-1][0],
				Size:  t.Asks[end-1][1],
			})
		}
	} else {
		for i := 0; i < len(t.Asks); i++ {
			if len(t.Asks[i]) < 2 {
				continue
			}
			r.Asks = append(r.Asks, global.DepthPair{
				Price: t.Asks[i][0],
				Size:  t.Asks[i][1],
			})
		}
	}

	// 买
	for i := 0; i < len(t.Bids); i++ {
		if len(t.Bids[i]) < 2 {
			continue
		}
		r.Bids = append(r.Bids, global.DepthPair{
			Price: t.Bids[i][0],
			Size:  t.Bids[i][1],
		})
	}
	return r, nil
}

// GetKline 获取k线数据
func (c *Client) GetKline(req global.KlineReq) ([]global.Kline, error) {
	groupSec := 60
	rangeHour := 1
	if req.Period == "5m" {
		groupSec = 300
		rangeHour = 12
	} else if req.Period == "15m" {
		groupSec = 900
		rangeHour = 24
	} else if req.Period == "30m" {
		groupSec = 1800
		rangeHour = 48
	} else if req.Period == "1h" {
		groupSec = 3600
		rangeHour = 96
	} else if req.Period == "8h" {
		groupSec = 28800
		rangeHour = 768
	} else if req.Period == "1d" {
		groupSec = 86400
		rangeHour = 2304
	}
	sym := strings.ToLower(req.Base + "_" + req.Quote)
	path := fmt.Sprintf("/api2/1/candlestick2/%s?group_sec=%d&range_hour=%d", sym, groupSec, rangeHour)
	rsp := struct {
		Result  string      `json:"result"`
		Message string      `json:"message"`
		Code    int64       `json:"code"`
		Data    [][]float64 `json:"data"`
	}{}

	e := c.httpReq("GET", path, nil, &rsp)
	if e != nil {
		return nil, e
	}
	if rsp.Result != "true" {
		return nil, fmt.Errorf(rsp.Message)
	}
	k := []global.Kline{}
	for i := 0; i < len(rsp.Data); i++ {
		if len(rsp.Data[i]) < 6 {
			fmt.Println("gate len(rsp.Data[i]) < 6")
			continue
		}
		k = append(k, global.Kline{
			Base:      req.Base,
			Quote:     req.Quote,
			Timestamp: int64(rsp.Data[i][0]),
			Volume:    rsp.Data[i][1],
			Close:     rsp.Data[i][2],
			High:      rsp.Data[i][3],
			Low:       rsp.Data[i][4],
			Open:      rsp.Data[i][5],
		})
	}
	return k, nil
}
