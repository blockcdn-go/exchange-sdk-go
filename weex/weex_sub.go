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

	sreq.Base, sreq.Quote = strings.ToUpper(sreq.Base), strings.ToUpper(sreq.Quote)
	symbol := sreq.Base + sreq.Quote
	req := struct {
		ID     int64    `json:"id"`
		Method string   `json:"method"`
		Params []string `json:"params"`
	}{ID: time.Now().Unix(), Method: "today.subscribe", Params: []string{}}
	req.Params = append(req.Params, symbol)
	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.sock.WriteJSON(req)
	if err != nil {
		log.Printf("发送消息失败 %s %s\n", symbol, err.Error())
		return nil, err
	}
	c.tick[sreq] = ch
	return ch, nil
}

// SubDepth 订阅深度行情，使用rest接口查询实现
func (c *Client) SubDepth(sreq global.TradeSymbol) (chan global.Depth, error) {
	in := make(map[string]interface{})
	in["market"] = strings.ToLower(sreq.Base + sreq.Quote)
	in["limit"] = 100
	in["merge"] = "0"

	ch := make(chan global.Depth, 100)
	go func() {
		for {
			d := struct {
				Asks [][]string `json:"asks"`
				Bids [][]string `json:"bids"`
			}{}
			r := weexRsp{Data: &d}
			err := c.httpReq("GET", "https://api.weex.com/v1/market/depth", in, &r, false)
			if err != nil || r.Code != 0 {
				log.Printf("weex get depth error: %+v, msg:%s\n", err, r.Msg)
				time.Sleep(10 * time.Second)
				continue
			}
			dr := global.Depth{
				Base:  sreq.Base,
				Quote: sreq.Quote,
				Asks:  []global.DepthPair{},
				Bids:  []global.DepthPair{},
			}
			for _, a := range d.Asks {
				if len(a) < 2 {
					continue
				}
				dr.Asks = append(dr.Asks, global.DepthPair{
					Price: toFloat(a[0]),
					Size:  toFloat(a[1]),
				})
			}
			for _, b := range d.Bids {
				if len(b) < 2 {
					continue
				}
				dr.Bids = append(dr.Bids, global.DepthPair{
					Price: toFloat(b[0]),
					Size:  toFloat(b[1]),
				})
			}
			ch <- dr
			time.Sleep(10 * time.Second)
		}
	}()

	return ch, nil
}

// SubLateTrade 查询交易详细数据
func (c *Client) SubLateTrade(sreq global.TradeSymbol) (chan global.LateTrade, error) {
	c.once.Do(func() { c.wsConnect() })
	if c.sock == nil {
		return nil, errors.New("wsconnect failed")
	}
	ch := make(chan global.LateTrade, 100)
	sreq.Base, sreq.Quote = strings.ToUpper(sreq.Base), strings.ToUpper(sreq.Quote)
	symbol := sreq.Base + sreq.Quote
	req := struct {
		ID     int64    `json:"id"`
		Method string   `json:"method"`
		Params []string `json:"params"`
	}{ID: time.Now().Unix(), Method: "deals.subscribe", Params: []string{}}
	req.Params = append(req.Params, symbol)
	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.sock.WriteJSON(req)
	if err != nil {
		log.Printf("发送消息失败 %s %s\n", symbol, err.Error())
		return nil, err
	}
	c.latetrade[sreq] = ch
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
		c.sock = nil
		return err
	}
	c.sock = conn

	//在这儿进行订阅消息重放
	if c.replay {
		log.Printf("连接成功，进行消息重放\n")
		for k := range c.tick {
			symbol := strings.ToUpper(k.Base + k.Quote)
			req := struct {
				ID     int64    `json:"id"`
				Method string   `json:"method"`
				Params []string `json:"params"`
			}{ID: time.Now().Unix(), Method: "today.subscribe", Params: []string{}}
			req.Params = append(req.Params, symbol)
			c.mutex.Lock()
			err := c.sock.WriteJSON(req)
			c.mutex.Unlock()
			if err != nil {
				log.Printf("订阅消息重放失败 %s %s\n", symbol, err.Error())
			}
		}

		//
		for k := range c.latetrade {
			symbol := strings.ToUpper(k.Base + k.Quote)
			req := struct {
				ID     int64    `json:"id"`
				Method string   `json:"method"`
				Params []string `json:"params"`
			}{ID: time.Now().Unix(), Method: "deals.subscribe", Params: []string{}}
			req.Params = append(req.Params, symbol)
			c.mutex.Lock()
			err := c.sock.WriteJSON(req)
			c.mutex.Unlock()
			if err != nil {
				log.Printf("订阅消息重放失败 %s %s\n", symbol, err.Error())
			}
		}
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
