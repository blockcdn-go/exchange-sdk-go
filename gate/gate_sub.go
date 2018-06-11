package gate

import (
	"fmt"
	"strings"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

// SubTicker ...
func (c *Client) SubTicker(sreq global.TradeSymbol) (chan global.Ticker, error) {
	sreq.Base = strings.ToUpper(sreq.Base)
	sreq.Quote = strings.ToUpper(sreq.Quote)
	ch := make(chan global.Ticker, 100)
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.tick[sreq] = ch

	c.tickonce.Do(func() { go c.loopTicker() })

	return ch, nil
}

// SubDepth 订阅深度行情
func (c *Client) SubDepth(sreq global.TradeSymbol) (chan global.Depth, error) {
	sreq.Base = strings.ToUpper(sreq.Base)
	sreq.Quote = strings.ToUpper(sreq.Quote)
	ch := make(chan global.Depth, 100)
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.depth[sreq] = ch

	c.depthonce.Do(func() { go c.loopDepth() })

	return ch, nil
}

func (c *Client) loopTicker() {
	for {
		r := map[string]TickerResponse{}
		err := c.httpReq("GET", "/api2/1/tickers", nil, &r)
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}
		// 遍历map得到base和quote进行行情分发
		for k, v := range r {
			base, quote := split(k)
			key := global.TradeSymbol{Base: base, Quote: quote}
			c.mutex.Lock()
			tch, ok := c.tick[key]
			c.mutex.Unlock()
			if ok {
				tch <- global.Ticker{
					Base:               base,
					Quote:              quote,
					PriceChange:        v.Last * (v.PercentChange / 100),
					PriceChangePercent: v.PercentChange,
					LastPrice:          v.Last,
					HighPrice:          v.High24hr,
					LowPrice:           v.Low24hr,
					Volume:             v.BaseVolume,
				}
			}
		}
		//
		time.Sleep(10 * time.Second)
	}
}

func (c *Client) loopDepth() {
	for {
		r := map[string]struct {
			Asks [][]float64 `json:"asks"` //卖方深度
			Bids [][]float64 `json:"bids"` //买方深度
		}{}
		err := c.httpReq("GET", "/api2/1/orderBooks", nil, &r)
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}

		for k, t := range r {
			base, quote := split(k)
			key := global.TradeSymbol{Base: base, Quote: quote}
			c.mutex.Lock()
			dch, ok := c.depth[key]
			c.mutex.Unlock()
			if !ok {
				continue
			}
			dp := global.Depth{
				Base:  base,
				Quote: quote,
				Asks:  []global.DepthPair{},
				Bids:  []global.DepthPair{},
			}

			if len(t.Asks) >= 2 && t.Asks[0][0] > t.Asks[1][0] {
				// 卖 倒序
				for end := len(t.Asks); end > len(t.Asks); end-- {
					if len(t.Asks[end-1]) < 2 {
						continue
					}
					dp.Asks = append(dp.Asks, global.DepthPair{
						Price: t.Asks[end-1][0],
						Size:  t.Asks[end-1][1],
					})
				}
			} else {
				for i := 0; i < len(t.Asks); i++ {
					if len(t.Asks[i]) < 2 {
						continue
					}
					dp.Asks = append(dp.Asks, global.DepthPair{
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
				dp.Bids = append(dp.Bids, global.DepthPair{
					Price: t.Bids[i][0],
					Size:  t.Bids[i][1],
				})
			}

			dch <- dp
		}
		//
		time.Sleep(10 * time.Second)
	}
}

// SubLateTrade 查询交易详细数据
func (c *Client) SubLateTrade(sreq global.TradeSymbol) (chan global.LateTrade, error) {
	sreq.Base = strings.ToUpper(sreq.Base)
	sreq.Quote = strings.ToUpper(sreq.Quote)
	ch := make(chan global.LateTrade, 100)

	go func() {
		for {
			symbol := strings.ToLower(sreq.Base) + "_" + strings.ToLower(sreq.Quote)
			path := fmt.Sprintf("/api2/1/tradeHistory/%s", symbol)
			rsp := struct {
				Result  string      `json:"result"`
				Message string      `json:"message"`
				Code    int64       `json:"code"`
				Data    []LateTrade `json:"data"`
			}{}

			e := c.httpReq("GET", path, nil, &rsp)
			if e != nil {
				time.Sleep(10 * time.Second)
				continue
			}
			if rsp.Result != "true" {
				time.Sleep(10 * time.Second)
				continue
			}
			for _, td := range rsp.Data {
				// 去重，不重复的进行推送，重复的不管
				if !c.findSameLateTrade(td) {
					ch <- global.LateTrade{
						Base:      sreq.Base,
						Quote:     sreq.Quote,
						DateTime:  td.DateTime,
						Num:       td.Num,
						Price:     td.Price,
						Dircetion: td.Dircetion,
						Total:     td.Total,
					}
				}
			}

			// 保存最近成交
			c.savelasttrade = rsp.Data
			time.Sleep(10 * time.Second)
		}
	}()
	return ch, nil
}

func (c *Client) findSameLateTrade(n LateTrade) bool {
	for _, ll := range c.savelasttrade {
		if ll == n {
			return true
		}
	}
	return false
}

func split(s string) (string, string) {
	bq := strings.Split(s, "_")
	if len(bq) != 2 {
		return s, "error"
	}
	return strings.ToUpper(bq[0]), strings.ToUpper(bq[1])
}
