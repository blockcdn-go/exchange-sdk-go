package huobi

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

func (c *Client) parse(msg []byte) {
	t := struct {
		CH     string `json:"ch"`
		Ticker struct {
			// ticker 数据
			Amount float64 `json:"amount"`
			Open   float64 `json:"open"`
			Close  float64 `json:"close"`
			High   float64 `json:"high"`
			Low    float64 `json:"low"`
			Vol    float64 `json:"vol"`
			// 最近成交数据
			Data []struct {
				Price     float64 `json:"price"`
				Time      int64   `json:"ts"`
				Amount    float64 `json:"amount"`
				Direction string  `json:"direction"`
			} `json:"data"`
			// 深度行情数据
			Asks [][]float64 `json:"asks"` //卖方深度
			Bids [][]float64 `json:"bids"` //买方深度
		} `json:"tick"`
	}{}
	//fmt.Println("huobipro: ", string(msg))
	err := json.Unmarshal(msg, &t)
	if err != nil {
		log.Printf("json unmarshal %s\n", err.Error())
		return
	}
	es := strings.Split(t.CH, ".")
	if len(es) < 3 {
		log.Println("huobipro ch error: ", es)
		return
	}

	// tick e.g "market.btcusdt.detail"
	// depth e.g "market.btcusdt.depth.step0"
	// latetrade e.g "market.btcusdt.trade.detail"
	base, quote := SplitSymbol(es[1])
	key := global.TradeSymbol{Base: base, Quote: quote}
	if es[2] == "detail" {
		c.mutex.Lock()
		ch, ok := c.tick[key]
		c.mutex.Unlock()
		if !ok {
			log.Printf("收到一个没有找到对应的消息 %+v %s\n", key, string(msg))
			return
		}
		v := t.Ticker.Close - t.Ticker.Open
		ret := global.Ticker{
			Base:               base,
			Quote:              quote,
			PriceChange:        v,
			PriceChangePercent: v / t.Ticker.Open * 100,
			LastPrice:          t.Ticker.Close,
			HighPrice:          t.Ticker.High,
			LowPrice:           t.Ticker.Low,
			Volume:             t.Ticker.Vol,
		}
		ch <- ret
	} else if es[2] == "depth" {
		c.mutex.Lock()
		ch, ok := c.depth[key]
		c.mutex.Unlock()
		if !ok {
			log.Printf("收到一个没有找到对应的消息 %+v %s\n", key, string(msg))
			return
		}
		ret := global.Depth{
			Base:  base,
			Quote: quote,
			Asks:  make([]global.DepthPair, 0, 5),
			Bids:  make([]global.DepthPair, 0, 5),
		}
		if len(t.Ticker.Asks) >= 2 && t.Ticker.Asks[0][0] > t.Ticker.Asks[1][0] {
			// 卖 倒序
			for end := len(t.Ticker.Asks); end > 0; end-- {
				ret.Asks = append(ret.Asks, global.DepthPair{
					Price: t.Ticker.Asks[end-1][0],
					Size:  t.Ticker.Asks[end-1][1]})
			}
		} else {
			for i := 0; i < len(t.Ticker.Asks); i++ {
				ret.Asks = append(ret.Asks, global.DepthPair{Price: t.Ticker.Asks[i][0],
					Size: t.Ticker.Asks[i][1]})
			}
		}

		// 买
		for i := 0; i < len(t.Ticker.Bids); i++ {
			ret.Bids = append(ret.Bids, global.DepthPair{Price: t.Ticker.Bids[i][0],
				Size: t.Ticker.Bids[i][1]})
		}
		ch <- ret
	} else if es[2] == "trade" {
		c.mutex.Lock()
		ch, ok := c.latetrade[key]
		c.mutex.Unlock()
		if !ok {
			log.Printf("收到一个没有找到对应的消息 %+v %s\n", key, string(msg))
			return
		}
		for _, d := range t.Ticker.Data {
			tm := time.Unix(d.Time/1000, 0)
			dt := tm.Format("2006-01-02 03:04:05 PM")
			lt := global.LateTrade{
				Base:      base,
				Quote:     quote,
				DateTime:  dt,
				Num:       d.Amount,
				Price:     d.Price,
				Dircetion: d.Direction,
				Total:     d.Price * d.Amount,
			}
			ch <- lt
		}
	}
}
