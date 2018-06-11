package zb

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/global"
	"github.com/blockcdn-go/exchange-sdk-go/utils"
	"github.com/gorilla/websocket"
)

// SubLateTrade 查询交易详细数据
func (c *Client) SubLateTrade(sreq global.TradeSymbol) (chan global.LateTrade, error) {
	c.otherOnce.Do(func() { c.wsOtherConnect() })
	if c.otherSock == nil {
		return nil, errors.New("ZB connect failed")
	}
	sreq.Base, sreq.Quote = strings.ToUpper(sreq.Base), strings.ToUpper(sreq.Quote)
	ch := make(chan global.LateTrade, 100)
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.latetrade[sreq] = ch
	req := struct {
		E string `json:"event"`
		C string `json:"channel"`
	}{E: "addChannel"}
	req.C = fmt.Sprintf("%s_trades", strings.ToLower(sreq.Base+sreq.Quote))
	c.otherSock.WriteJSON(req)
	return ch, nil
}

// SubTicker ...
func (c *Client) SubTicker(sreq global.TradeSymbol) (chan global.Ticker, error) {
	c.tickOnce.Do(func() { c.wsTickerConnect() })
	if c.tickSock == nil {
		return nil, errors.New("ZB connect failed")
	}
	sreq.Base, sreq.Quote = strings.ToUpper(sreq.Base), strings.ToUpper(sreq.Quote)
	ch := make(chan global.Ticker, 100)
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.tick[sreq] = ch

	return ch, nil
}

// SubDepth 订阅深度行情
func (c *Client) SubDepth(sreq global.TradeSymbol) (chan global.Depth, error) {
	c.otherOnce.Do(func() { c.wsOtherConnect() })
	if c.otherSock == nil {
		return nil, errors.New("ZB connect failed")
	}
	sreq.Base, sreq.Quote = strings.ToUpper(sreq.Base), strings.ToUpper(sreq.Quote)
	ch := make(chan global.Depth, 100)
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.depth[sreq] = ch
	req := struct {
		E string `json:"event"`
		C string `json:"channel"`
	}{E: "addChannel"}
	req.C = fmt.Sprintf("%s_depth", strings.ToLower(sreq.Base+sreq.Quote))
	c.otherSock.WriteJSON(req)
	return ch, nil
}

func (c *Client) connect(wsaddr string) (*websocket.Conn, error) {
	log.Printf("ZB 连接 %s 中... ", wsaddr)
	conn, _, err := c.config.WSSDialer.Dial(wsaddr, nil)
	log.Printf("连接: %s\n", utils.Ternary(err == nil, "成功", "失败").(string))
	return conn, err
}

func (c *Client) wsTickerConnect() error {
	c.tickSock = nil
	conn, err := c.connect("wss://kline.zb.com:2443/websocket")
	if err != nil {
		log.Printf("zb <wss://kline.zb.com:2443/websocket> connect failed %+v\n", err)
		c.tickSock = nil
		return err
	}
	c.tickSock = conn

	//订阅所有ticker
	req := struct {
		E string `json:"event"`
		C string `json:"channel"`
		B string `json:"binary"`
		Z string `json:"isZip"`
	}{E: "addChannel", B: "false", Z: "false"}
	req.C = "top_all_qc"
	c.tickSock.WriteJSON(req)
	req.C = "top_all_zb"
	c.tickSock.WriteJSON(req)
	req.C = "top_all_usdt"
	c.tickSock.WriteJSON(req)
	req.C = "top_all_btc"
	c.tickSock.WriteJSON(req)

	// 循环读取消息
	go func() {
		for {
			msg, err := c.readWSMessage(c.tickSock)
			if err != nil {
				log.Printf("ZB < %s > 断开连接，五秒后重连...\n", err.Error())
				go func() {
					time.Sleep(5 * time.Second)
					c.wsTickerConnect()
				}()
				return
			}
			if msg == nil {
				continue
			}

			// 业务逻辑处理
			c.parse(msg)

		}

	}()
	return err
}

func (c *Client) wsOtherConnect() error {
	c.otherSock = nil
	conn, err := c.connect("wss://api.zb.com:9999/websocket")
	if err != nil {
		log.Printf("zb <wss://api.zb.com:9999/websocket> connect failed %+v\n", err)
		c.otherSock = nil
		return err
	}
	c.otherSock = conn

	if c.replay {
		req := struct {
			E string `json:"event"`
			C string `json:"channel"`
		}{E: "addChannel"}
		c.mutex.Lock()
		for k := range c.depth {
			req.C = fmt.Sprintf("%s_depth", strings.ToLower(k.Base+k.Quote))
			c.otherSock.WriteJSON(req)
		}
		for k := range c.latetrade {
			req.C = fmt.Sprintf("%s_trades", strings.ToLower(k.Base+k.Quote))
			c.otherSock.WriteJSON(req)
		}
		c.mutex.Unlock()
	}
	c.replay = true
	// 循环读取消息
	go func() {
		for {
			msg, err := c.readWSMessage(c.otherSock)
			if err != nil {
				log.Printf("ZB < %s > 断开连接，五秒后重连...\n", err.Error())
				go func() {
					time.Sleep(5 * time.Second)
					c.wsOtherConnect()
				}()
				return
			}
			if msg == nil {
				continue
			}

			// 业务逻辑处理
			c.parse(msg)

		}

	}()

	return nil
}

func (c *Client) readWSMessage(conn *websocket.Conn) ([]byte, error) {
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	return msg, nil
}
