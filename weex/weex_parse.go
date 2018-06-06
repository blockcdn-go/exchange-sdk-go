package weex

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

func (c *Client) parse(msg []byte) {
	r := struct {
		Method string        `json:"method"`
		Params []interface{} `json:"params"`
	}{Params: []interface{}{}}

	err := json.Unmarshal(msg, &r)
	if err != nil || r.Method == "" || len(r.Params) == 0 {
		return
	}
	len := (len(r.Params) / 2) * 2

	for i := 0; i < len; i += 2 {
		base, quote := split(r.Params[i].(string))
		key := global.TradeSymbol{Base: base, Quote: quote}
		if strings.Contains(r.Method, "today") {
			c.mutex.Lock()
			ch, ok := c.tick[key]
			c.mutex.Unlock()
			if !ok {
				log.Printf("收到一个没有找到对应的消息 %+v %s\n", key, string(msg))
				return
			}
			v1m := r.Params[i+1].(map[string]interface{})
			open := toFloat(v1m["open"])
			ret := global.Ticker{
				Base:      base,
				Quote:     quote,
				LastPrice: toFloat(v1m["last"]),
				HighPrice: toFloat(v1m["high"]),
				LowPrice:  toFloat(v1m["low"]),
				Volume:    toFloat(v1m["volume"]),
			}
			v := ret.LastPrice - open
			ret.PriceChange = v / open * 100
			ch <- ret
		} else if strings.Contains(r.Method, "deals") {
			c.mutex.Lock()
			ch, ok := c.latetrade[key]
			c.mutex.Unlock()
			if !ok {
				log.Printf("收到一个没有找到对应的消息 %+v %s\n", key, string(msg))
				return
			}
			sl := r.Params[i+1].([]interface{})
			for _, s := range sl {
				v1m := s.(map[string]interface{})
				tm := time.Unix(int64(toFloat(v1m["time"])), 0)
				dt := tm.Format("2006-01-02 03:04:05 PM")
				lt := global.LateTrade{
					Base:      base,
					Quote:     quote,
					DateTime:  dt,
					Num:       toFloat(v1m["amount"]),
					Price:     toFloat(v1m["price"]),
					Dircetion: toString(v1m["type"]),
				}
				lt.Total = lt.Price * lt.Num
				ch <- lt
			}
		}
	}

}

func split(symbol string) (string, string) {
	r1 := symbol
	r2 := "error"
	if len(symbol) < 5 {
		return r1, r2
	}

	b := []byte(symbol)

	l3 := string(b[len(b)-3 : len(b)])
	if strings.ToUpper(l3) == "BTC" {
		r1 = strings.ToUpper(string(b[0 : len(b)-3]))
		r2 = "BTC"
	}
	if strings.ToUpper(l3) == "ETH" {
		r1 = strings.ToUpper(string(b[0 : len(b)-3]))
		r2 = "ETH"
	}
	if strings.ToUpper(l3) == "USD" {
		r1 = strings.ToUpper(string(b[0 : len(b)-3]))
		r2 = "USD"
	}
	return r1, r2
}
