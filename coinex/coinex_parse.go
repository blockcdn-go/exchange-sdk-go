package coinex

import (
	"encoding/json"
	"strings"

	"github.com/blockcdn-go/exchange-sdk-go/global"

	"github.com/blockcdn-go/exchange-sdk-go/utils"
)

func (c *Client) parse(msg []byte) {
	r := map[string]interface{}{}
	err := json.Unmarshal(msg, &r)
	if err != nil {
		return
	}

	method := utils.ToString(r["method"])
	params, ok := r["params"]
	if !ok {
		return
	}
	if strings.Contains(method, "state.update") {
		ticks := params.([]interface{})
		for _, arr := range ticks {
			t := arr.(map[string]interface{})
			for k, v := range t {
				ticker := v.(map[string]interface{})
				base, quote := split(utils.ToString(k))
				key := global.TradeSymbol{Base: base, Quote: quote}
				c.mtx.Lock()
				ch, ok := c.tick[key]
				c.mtx.Unlock()
				if !ok {
					return
				}
				open := utils.ToFloat(ticker["open"])
				ret := global.Ticker{
					Base:      base,
					Quote:     quote,
					LastPrice: utils.ToFloat(ticker["last"]),
					HighPrice: utils.ToFloat(ticker["high"]),
					LowPrice:  utils.ToFloat(ticker["low"]),
					Volume:    utils.ToFloat(ticker["volume"]),
				}
				if ret.LastPrice == 0. {
					return
				}
				if open != 0. {
					ret.PriceChange = ret.LastPrice - open
					ret.PriceChangePercent = ret.PriceChange / ret.LastPrice * 100
				}
				ch <- ret
			}

		}
	}

}
