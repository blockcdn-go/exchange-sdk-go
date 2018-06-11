package zb

import "github.com/blockcdn-go/exchange-sdk-go/global"

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
	return nil, err
}
