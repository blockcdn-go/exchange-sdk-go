package bitstamp

import (
	"github.com/blockcdn-go/exchange-sdk-go/global"
	"github.com/blockcdn-go/exchange-sdk-go/utils"
)

// GetAllSymbol 交易市场详细行情接口
func (c *Client) GetAllSymbol() ([]global.TradeSymbol, error) {
	r := []map[string]interface{}{}
	err := c.httpReq("GET", "https://www.bitstamp.net/api/v2/trading-pairs-info/", nil, &r, false)
	if err != nil {
		return nil, err
	}

	ret := []global.TradeSymbol{}
	for _, s := range r {
		base, quote := split(utils.ToString(s["name"]))
		ret = append(ret, global.TradeSymbol{
			Base:  base,
			Quote: quote,
		})
	}

	return ret, nil
}
