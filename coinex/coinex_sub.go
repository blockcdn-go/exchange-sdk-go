package coinex

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/blockcdn-go/exchange-sdk-go/global"
	"github.com/blockcdn-go/exchange-sdk-go/utils"
)

// SubTicker ...
func (c *Client) SubTicker(sreq global.TradeSymbol) (chan global.Ticker, error) {
	c.tickOnce.Do(func() { c.wsConnect("wss://socket.coinex.com/") })
	if c.sock == nil {
		return nil, errors.New("connect fialed")
	}
	ch := make(chan global.Ticker, 100)
	sreq.Base, sreq.Quote = strings.ToUpper(sreq.Base), strings.ToUpper(sreq.Quote)
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.tick[sreq] = ch

	return ch, nil
}

func (c *Client) connect(wsaddr string) (*websocket.Conn, error) {
	log.Printf("coinex 连接 %s 中... ", wsaddr)
	conn, _, err := c.Config.WSSDialer.Dial(wsaddr, nil)
	log.Printf("连接: %s\n", utils.Ternary(err == nil, "成功", "失败").(string))
	return conn, err
}

func (c *Client) wsConnect(wsaddr string) {
	sock, err := c.connect(wsaddr)
	if err != nil {
		log.Printf("coinex connect failed %+v\n", err)
		c.sock = nil
		return
	}
	c.sock = sock

	req := map[string]interface{}{}
	req["id"] = time.Now().Unix()
	req["method"] = "state.subscribe"
	req["params"] = []interface{}{}
	c.sock.WriteJSON(req)

	go func() {
		ping := time.NewTicker(time.Second * 10)
		for {
			select {
			case <-ping.C:
				ping = time.NewTicker(time.Second * 10)
				break
			default:
				msg, err := c.readWSMessage(c.sock)
				if err != nil {
					log.Printf("coinex < %s > 断开连接，五秒后重连...\n", err.Error())
					go func() {
						time.Sleep(5 * time.Second)
						c.wsConnect(wsaddr)
					}()
					return
				}
				if msg == nil {
					continue
				}

				// 业务逻辑处理
				c.parse(msg)
				fmt.Println(string(msg))
			}
		}
	}()
}

func (c *Client) readWSMessage(conn *websocket.Conn) ([]byte, error) {
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	if strings.Contains(string(msg), "pong") {
		return nil, nil
	}
	return msg, nil
}
