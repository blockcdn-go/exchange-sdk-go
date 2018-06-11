package zb

import (
	"fmt"
	"strings"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/utils"

	"github.com/blockcdn-go/exchange-sdk-go/global"

	jsoniter "github.com/json-iterator/go"
)

func (c *Client) parse(msg []byte) {
	data := map[string]interface{}{}
	if msg[0] == '(' {
		msg = msg[1:]
	}
	if msg[len(msg)-1] == ')' {
		msg = msg[0 : len(msg)-1]
	}
	err := jsoniter.Unmarshal(msg, &data)
	if err != nil {
		fmt.Printf("json error: %+v\n", err)
		return
	}
	dtype, _ := data["dataType"].(string)
	if strings.Contains(dtype, "topAll") {
		// ticker
		darr := data["datas"].([]interface{})
		for _, da := range darr {
			t := da.(map[string]interface{})
			base, quote := split(t["market"].(string))
			key := global.TradeSymbol{Base: base, Quote: quote}
			c.mutex.Lock()
			ch, ok := c.tick[key]
			c.mutex.Unlock()
			if !ok {
				continue
			}
			ret := global.Ticker{
				Base:               base,
				Quote:              quote,
				LastPrice:          utils.ToFloat(t["lastPrice"]),
				HighPrice:          utils.ToFloat(t["hightPrice"]),
				LowPrice:           utils.ToFloat(t["lowPrice"]),
				Volume:             utils.ToFloat(t["vol"]),
				PriceChangePercent: utils.ToFloat(t["riseRate"]),
			}
			ret.PriceChange = ret.LastPrice * (ret.PriceChangePercent / 100)
			ch <- ret
		}
	} else if strings.Contains(dtype, "depth") {
		base, quote := split2(strings.ToUpper(data["channel"].(string)))
		key := global.TradeSymbol{Base: base, Quote: quote}
		c.mutex.Lock()
		ch, ok := c.depth[key]
		c.mutex.Unlock()
		if !ok {
			return
		}
		asks := []interface{}{}
		bids := []interface{}{}
		ret := global.Depth{
			Base:  base,
			Quote: quote,
			Asks:  []global.DepthPair{},
			Bids:  []global.DepthPair{},
		}

		if _, ok := data["asks"]; ok {
			asks = data["asks"].([]interface{})
			for _, a := range asks {
				aa := a.([]interface{})
				if len(aa) < 2 {
					continue
				}
				ret.Asks = append(ret.Asks, global.DepthPair{
					Price: utils.ToFloat(aa[0]),
					Size:  utils.ToFloat(aa[1]),
				})
			}

		}
		if _, ok := data["bids"]; ok {
			bids = data["bids"].([]interface{})
			for _, b := range bids {
				bb := b.([]interface{})
				if len(bb) < 2 {
					continue
				}
				ret.Bids = append(ret.Bids, global.DepthPair{
					Price: utils.ToFloat(bb[0]),
					Size:  utils.ToFloat(bb[1]),
				})
			}
		}

		ch <- ret

	} else if strings.Contains(dtype, "trades") {
		ds, e := data["data"]
		if !e {
			return
		}
		base, quote := split2(strings.ToUpper(data["channel"].(string)))
		key := global.TradeSymbol{Base: base, Quote: quote}
		c.mutex.Lock()
		ch, ok := c.latetrade[key]
		c.mutex.Unlock()
		if !ok {
			return
		}
		dd := ds.([]interface{})
		for _, lts := range dd {
			m := lts.(map[string]interface{})
			tm := time.Unix(int64(utils.ToFloat(m["date"])), 0)
			dt := tm.Format("2006-01-02 03:04:05 PM")
			lt := global.LateTrade{
				Base:      base,
				Quote:     quote,
				DateTime:  dt,
				Num:       utils.ToFloat(m["amount"]),
				Price:     utils.ToFloat(m["price"]),
				Dircetion: utils.ToString(m["type"]),
			}
			lt.Total = lt.Price * lt.Num

			ch <- lt
		}

	} else {
		fmt.Println(string(msg))
	}
}
