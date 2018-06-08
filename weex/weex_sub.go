package weex

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/global"
	"github.com/gorilla/websocket"
)

// SubTicker ...
func (c *Client) SubTicker(sreq global.TradeSymbol) (chan global.Ticker, error) {
	c.once.Do(func() { c.wsConnect() })
	if c.sock == nil {
		return nil, errors.New("wsconnect failed")
	}
	ch := make(chan global.Ticker, 100)
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.tick[sreq] = ch

	sreq.Base, sreq.Quote = strings.ToUpper(sreq.Base), strings.ToUpper(sreq.Quote)

	req := struct {
		ID     int64    `json:"id"`
		Method string   `json:"method"`
		Params []string `json:"params"`
	}{ID: time.Now().Unix(), Method: "today.subscribe", Params: []string{}}
	for k := range c.tick {
		req.Params = append(req.Params, strings.ToUpper(k.Base+k.Quote))
	}

	err := c.sock.WriteJSON(req)
	if err != nil {
		log.Printf("发送消息失败 %+v %s\n", req, err.Error())
		return nil, err
	}

	return ch, nil
}

// SubDepth 订阅深度行情
func (c *Client) SubDepth(sreq global.TradeSymbol) (chan global.Depth, error) {
	con, err := c.connect()
	if con == nil {
		return nil, err
	}
	sreq.Base, sreq.Quote = strings.ToUpper(sreq.Base), strings.ToUpper(sreq.Quote)

	c.mutex.Lock()
	defer c.mutex.Unlock()
	ch, ok := c.depth[sreq]
	if !ok {
		ch = make(chan global.Depth, 100)
		c.depth[sreq] = ch
	}

	req := struct {
		ID     int64         `json:"id"`
		Method string        `json:"method"`
		Params []interface{} `json:"params"`
	}{ID: time.Now().Unix(), Method: "depth.subscribe", Params: []interface{}{}}
	req.Params = append(req.Params, sreq.Base+sreq.Quote, 20, "0")

	err = con.WriteJSON(req)
	if err != nil {
		log.Printf("发送消息失败 %+v %s\n", req, err.Error())
		return nil, err
	}

	go c.depthLoop(sreq, con)
	return ch, nil
}

func (c *Client) depthLoop(sreq global.TradeSymbol, con *websocket.Conn) {
	for {
		msg, err := c.readWSMessage(con)
		if err != nil {
			for {
				log.Printf("weex disconnect %+v, reconnect after five seconds\n", err)
				time.Sleep(5 * time.Second)
				_, err := c.SubDepth(sreq)
				if err != nil {
					break
				}
			}
			return
		}
		c.parse(msg)
	}
}

// SubLateTrade 查询交易详细数据
func (c *Client) SubLateTrade(sreq global.TradeSymbol) (chan global.LateTrade, error) {
	c.once.Do(func() { c.wsConnect() })
	if c.sock == nil {
		return nil, errors.New("wsconnect failed")
	}
	ch := make(chan global.LateTrade, 100)
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.latetrade[sreq] = ch

	sreq.Base, sreq.Quote = strings.ToUpper(sreq.Base), strings.ToUpper(sreq.Quote)
	req := struct {
		ID     int64    `json:"id"`
		Method string   `json:"method"`
		Params []string `json:"params"`
	}{ID: time.Now().Unix(), Method: "deals.subscribe", Params: []string{}}
	for k := range c.depth {
		req.Params = append(req.Params, strings.ToUpper(k.Base+k.Quote))
	}

	err := c.sock.WriteJSON(req)
	if err != nil {
		log.Printf("发送消息失败 %+v %s\n", req, err.Error())
		return nil, err
	}

	return ch, nil
}

func (c *Client) connect() (*websocket.Conn, error) {
	wsaddr := "wss://ws.weexpro.com/"
	log.Printf("huobi 连接 %s 中...\n", wsaddr)
	conn, _, err := c.config.WSSDialer.Dial(wsaddr, nil)
	return conn, err
}

func (c *Client) wsConnect() error {
	c.sock = nil
	conn, err := c.connect()
	if err != nil {
		log.Printf("weex connect failed %+v\n", err)
		c.sock = nil
		return err
	}
	c.sock = conn

	//在这儿进行订阅消息重放
	if c.replay {
		log.Printf("连接成功，进行消息重放\n")
		req := struct {
			ID     int64    `json:"id"`
			Method string   `json:"method"`
			Params []string `json:"params"`
		}{ID: time.Now().Unix(), Method: "today.subscribe", Params: []string{}}
		c.mutex.Lock()
		for k := range c.tick {
			req.Params = append(req.Params, strings.ToUpper(k.Base+k.Quote))
		}
		err := c.sock.WriteJSON(req)
		if err != nil {
			log.Printf("订阅消息重放失败 %+v %s\n", req, err.Error())
		}

		req.Params = []string{}
		req.Method = "deals.subscribe"
		for k := range c.latetrade {
			req.Params = append(req.Params, strings.ToUpper(k.Base+k.Quote))
		}
		err = c.sock.WriteJSON(req)
		if err != nil {
			log.Printf("订阅消息重放失败 %+v %s\n", req, err.Error())
		}

		c.mutex.Unlock()

	}
	c.replay = true
	// 循环读取消息
	go func() {
		for {
			msg, err := c.readWSMessage(c.sock)
			if err != nil {
				log.Printf("huobipro < %s > 断开连接，五秒后重连...\n", err.Error())
				go func() {
					time.Sleep(5 * time.Second)
					c.wsConnect()
				}()
				return
			}
			if msg == nil {
				continue
			}

			// 业务逻辑处理
			c.parse(msg)
			//fmt.Println(string(msg))
		}

	}()
	return err
}

func (c *Client) readWSMessage(conn *websocket.Conn) ([]byte, error) {
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	if strings.Contains(string(msg), "ping") {
		//fmt.Println(string(message))
		var ping struct {
			Ping int64 `json:"ping"`
		}
		err := json.Unmarshal(msg, &ping)
		if err != nil {
			return nil, nil
		}
		pong := struct {
			Pong int64 `json:"pong"`
		}{ping.Ping}
		conn.WriteJSON(pong)
		//fmt.Printf("%+v\n", pong)
		return nil, nil
	}
	return msg, nil
}
