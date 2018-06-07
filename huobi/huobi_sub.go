package huobi

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

// GetKline websocket 查询kline
func (c *Client) GetKline(req global.KlineReq) ([]global.Kline, error) {
	conn, err := c.connect()
	if err != nil {
		return nil, err
	}
	period := req.Period
	if strings.Contains(period, "m") {
		period = period + "in"
	} else if period == "1h" {
		period = "60m"
	} else if strings.Contains(period, "d") {
		period = period + "ay"
	} else if strings.Contains(period, "w") {
		period = period + "eek"
	}
	symbol := strings.ToLower(req.Base + req.Quote)
	topic := fmt.Sprintf("market.%s.kline.%s", symbol, period)
	kreq := struct {
		Topic string `json:"req"`
		ID    string `json:"id"`
		From  int64  `json:"from,omitempty"`
		To    int64  `json:"to,omitempty"`
	}{Topic: topic, ID: c.generateClientID()}

	c.mutex.Lock()
	err = conn.WriteJSON(kreq)
	c.mutex.Unlock()
	if err != nil {
		return nil, err
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(msg)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	message, _ := ioutil.ReadAll(gz)
	rsp := struct {
		Status string  `json:"status"`
		Data   []Kline `json:"data"`
		Errmsg string  `json:"err-msg"`
	}{}
	err = json.Unmarshal(message, &rsp)
	if err != nil {
		return nil, err
	}
	if rsp.Status != "ok" {
		return nil, errors.New("huobipro websocket kline error:" + rsp.Errmsg)
	}
	ik := []global.Kline{}
	for _, k := range rsp.Data {
		ik = append(ik, global.Kline{
			Base:      k.Base,
			Quote:     k.Quote,
			Timestamp: int64(k.Timestamp),
			High:      k.High,
			Open:      k.Open,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
		})
	}
	return ik, nil
}

// SubDepth 查询市场深度数据
// type 可选值：{ step0, step1, step2, step3, step4, step5 } （合并深度0-5）；
// step0时，不合并深度
func (c *Client) SubDepth(sreq global.TradeSymbol) (chan global.Depth, error) {
	c.once.Do(func() { c.wsConnect() })
	if c.sock == nil {
		return nil, errors.New("connect failed")
	}
	sreq.Base = strings.ToUpper(sreq.Base)
	sreq.Quote = strings.ToUpper(sreq.Quote)
	symbol := strings.ToLower(sreq.Base + sreq.Quote)

	topic := fmt.Sprintf("market.%s.depth.%s", symbol, "step0")
	req := struct {
		Topic string `json:"sub"`
		ID    string `json:"id"`
	}{topic, c.generateClientID()}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	err := c.sock.WriteJSON(req)
	if err != nil {
		return nil, err
	}

	ch := make(chan global.Depth, 100)
	c.depth[sreq] = ch

	// 直接返回
	return ch, nil
}

// SubLateTrade 查询交易详细数据
func (c *Client) SubLateTrade(sreq global.TradeSymbol) (chan global.LateTrade, error) {
	c.once.Do(func() { c.wsConnect() })
	if c.sock == nil {
		return nil, errors.New("connect failed")
	}
	sreq.Base = strings.ToUpper(sreq.Base)
	sreq.Quote = strings.ToUpper(sreq.Quote)
	symbol := strings.ToLower(sreq.Base + sreq.Quote)

	topic := fmt.Sprintf("market.%s.trade.detail", symbol)
	req := struct {
		Topic string `json:"sub"`
		ID    string `json:"id"`
	}{topic, c.generateClientID()}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	err := c.sock.WriteJSON(req)
	if err != nil {
		return nil, err
	}
	ch := make(chan global.LateTrade, 100)
	c.latetrade[sreq] = ch

	return ch, nil
}

// SubTicker ...
func (c *Client) SubTicker(sreq global.TradeSymbol) (chan global.Ticker, error) {
	c.once.Do(func() { c.wsConnect() })
	if c.sock == nil {
		return nil, errors.New("connect failed")
	}
	sreq.Base = strings.ToUpper(sreq.Base)
	sreq.Quote = strings.ToUpper(sreq.Quote)
	symbol := strings.ToLower(sreq.Base + sreq.Quote)
	topic := fmt.Sprintf("market.%s.detail", symbol)
	req := struct {
		Topic string `json:"sub"`
		ID    string `json:"id"`
	}{topic, c.generateClientID()}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.sock.WriteJSON(req)
	if err != nil {
		return nil, err
	}
	ch := make(chan global.Ticker, 100)
	c.tick[sreq] = ch

	return ch, nil
}

func (c *Client) connect() (*websocket.Conn, error) {
	u := url.URL{Scheme: "wss", Host: *c.config.WSSHost, Path: "/ws"}
	log.Printf("huobi 连接 %s 中...\n", u.String())
	conn, _, err := c.config.WSSDialer.Dial(u.String(), nil)
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
			symbol := strings.ToLower(k.Base + k.Quote)
			topic := fmt.Sprintf("market.%s.detail", symbol)
			req := struct {
				Topic string `json:"sub"`
				ID    string `json:"id"`
			}{topic, c.generateClientID()}
			c.mutex.Lock()
			err := c.sock.WriteJSON(req)
			c.mutex.Unlock()
			if err != nil {
				log.Printf("订阅消息重放失败 %s %s\n", topic, err.Error())
			}
		}

		//
		for k := range c.depth {
			symbol := strings.ToLower(k.Base + k.Quote)
			topic := fmt.Sprintf("market.%s.depth.%s", symbol, "step0")
			req := struct {
				Topic string `json:"sub"`
				ID    string `json:"id"`
			}{topic, c.generateClientID()}
			c.mutex.Lock()
			err := c.sock.WriteJSON(req)
			c.mutex.Unlock()
			if err != nil {
				log.Printf("订阅消息重放失败 %s %s\n", topic, err.Error())
			}
		}

		//
		for k := range c.latetrade {
			symbol := strings.ToLower(k.Base + k.Quote)
			topic := fmt.Sprintf("market.%s.trade.detail", symbol)
			req := struct {
				Topic string `json:"sub"`
				ID    string `json:"id"`
			}{topic, c.generateClientID()}
			c.mutex.Lock()
			err := c.sock.WriteJSON(req)
			c.mutex.Unlock()
			if err != nil {
				log.Printf("订阅消息重放失败 %s %s\n", topic, err.Error())
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

		}

	}()
	return err
}

func (c *Client) readWSMessage(conn *websocket.Conn) ([]byte, error) {
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(msg)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return nil, nil
	}
	message, _ := ioutil.ReadAll(gz)

	if strings.Contains(string(message), "ping") {
		//fmt.Println(string(message))
		var ping struct {
			Ping int64 `json:"ping"`
		}
		err := json.Unmarshal(message, &ping)
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
	return message, nil
}
