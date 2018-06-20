package baseclass

import (
	"fmt"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

// SubTicker ...
func (c *Client) SubTicker(sreq global.TradeSymbol) (chan global.Ticker, error) {
	ch := make(chan global.Ticker, 100)
	//启动协程轮询
	go func() {
		for {
			t, err := c.Client.AicoinGetTicker(c.Exchange, sreq)
			if err != nil {
				fmt.Println(c.Exchange, " ticker error: ", err.Error())
			} else {
				ch <- t
			}
			time.Sleep(10 * time.Second)
		}
	}()
	return ch, nil
}

// SubDepth 订阅深度行情
func (c *Client) SubDepth(sreq global.TradeSymbol) (chan global.Depth, error) {
	ch := make(chan global.Depth, 100)
	//启动协程轮询
	go func() {
		for {
			t, err := c.Client.AicoinGetDepth(c.Exchange, sreq)
			if err != nil {
				fmt.Println(c.Exchange, " depth error: ", err.Error())
			} else {
				ch <- t
			}
			time.Sleep(10 * time.Second)
		}
	}()
	return ch, nil
}

// SubLateTrade 订阅交易详细数据
func (c *Client) SubLateTrade(sreq global.TradeSymbol) (chan global.LateTrade, error) {
	ch := make(chan global.LateTrade, 100)
	//启动协程轮询
	go func() {
		for {
			t, err := c.Client.AicoinGetLateTrade(c.Exchange, sreq)
			if err != nil {
				fmt.Println(c.Exchange, " latetrade error: ", err.Error())
			} else {
				c.mutex.Lock()
				lt := c.lastt[sreq]
				// 保存数据
				c.lastt[sreq] = t
				c.mutex.Unlock()

				// 推送
				for _, l := range t {
					find := false
					for _, it := range lt {
						if it == l {
							find = true
							break
						}
					}
					if !find {
						ch <- l
					}
				}
			}
			time.Sleep(10 * time.Second)
		}
	}()
	return ch, nil
}
