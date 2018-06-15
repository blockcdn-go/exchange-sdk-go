package coinegg

import (
	"fmt"
	"log"
	"strings"

	"github.com/blockcdn-go/exchange-sdk-go/utils"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

type errInfo struct {
	Code   int  `json:"code"`
	Result bool `json:"result"`
}

// GetAllSymbol 交易市场详细行情接口
func (c *Client) GetAllSymbol() ([]global.TradeSymbol, error) {
	ss := []global.TradeSymbol{}
	qs := []string{"btc", "eth", "usdt", "usc"}
	for _, q := range qs {
		r := map[string]interface{}{}
		path := fmt.Sprintf("https://www.coinegg.com/coin/%s/allcoin", q)
		err := c.httpReq("GET", path, nil, &r, false)
		if err != nil {
			log.Printf("%s error: %s\n", path, err.Error())
			continue
		}
		for k := range r {
			ss = append(ss, global.TradeSymbol{
				Base:  strings.ToUpper(k),
				Quote: strings.ToUpper(q),
			})
		}
	}

	return ss, nil
}

// GetDepth 获取深度行情
func (c *Client) GetDepth(req global.TradeSymbol) (global.Depth, error) {
	path := fmt.Sprintf("https://www.coinegg.com/coin/%s/%s/tradelist", req.Quote, req.Base)
	r := struct {
		errInfo
		Bids [][]interface{} `json:"buy"`
		Asks [][]interface{} `json:"sell"`
	}{}
	err := c.httpReq("GET", strings.ToLower(path), nil, &r, false)
	if err != nil {
		return global.Depth{}, err
	}
	if r.errInfo.Code != 0 {
		return global.Depth{}, fmt.Errorf("error code: %d", r.errInfo.Code)
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
		dp.Asks = append(dp.Asks, global.DepthPair{
			Price: utils.ToFloat(a[0]),
			Size:  utils.ToFloat(a[1]),
		})
	}
	for _, b := range r.Bids {
		if len(b) < 2 {
			continue
		}
		dp.Bids = append(dp.Bids, global.DepthPair{
			Price: utils.ToFloat(b[0]),
			Size:  utils.ToFloat(b[1]),
		})
	}
	return dp, nil
}

// GetKline 获取k线数据
func (c *Client) GetKline(req global.KlineReq) ([]global.Kline, error) {
	return nil, nil
}
