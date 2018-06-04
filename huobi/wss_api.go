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

// SubMarketKLine 查询市场K线图
// period 可选 1min, 5min, 15min, 30min, 60min, 1day, 1mon, 1week, 1year
func (c *WSSClient) SubMarketKLine(symbol string, period string) (<-chan []byte, error) {
	cid, conn, err := c.connect()
	if err != nil {
		return nil, err
	}

	topic := fmt.Sprintf("market.%s.kline.%s", symbol, period)
	req := struct {
		Topic string `json:"sub"`
		ID    string `json:"id"`
	}{topic, c.generateClientID()}

	err = conn.WriteJSON(req)
	if err != nil {
		c.Close()
		return nil, err
	}

	result := make(chan []byte)
	go c.start(topic, cid, result)
	return result, nil
}

// GetKline websocket 查询kline
func (c *WSSClient) GetKline(req global.KlineReq) ([]global.Kline, error) {
	_, conn, err := c.connect()
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

	err = conn.WriteJSON(kreq)
	if err != nil {
		c.Close()
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
func (c *WSSClient) SubDepth(sreq global.TradeSymbol) (chan global.Depth, error) {
	conn, err := c.wsConnect()
	if err != nil {
		return nil, err
	}
	symbol := strings.ToLower(sreq.Base + sreq.Quote)

	topic := fmt.Sprintf("market.%s.depth.%s", symbol, "step0")
	req := struct {
		Topic string `json:"sub"`
		ID    string `json:"id"`
	}{topic, c.generateClientID()}

	err = conn.WriteJSON(req)
	if err != nil {
		c.Close()
		return nil, err
	}

	ch := make(chan global.Depth, 100)

	go func() {
		for {
			msg, err := c.readWSMessage(conn)
			if err != nil {
				log.Printf("SubDepth websocket 连接断开 %s\n", err.Error())
				go func() {
					for {
						time.Sleep(5 * time.Second)
						ch, err = c.SubDepth(sreq)
						if err == nil {
							log.Println("重新连接成功...")
							return
						}
					}
				}()
				return
			}
			if msg == nil {
				continue
			}
			// 解析数据

			t := struct {
				Ticker struct {
					Asks [][]float64 `json:"asks"` //卖方深度
					Bids [][]float64 `json:"bids"` //买方深度
				} `json:"tick"`
			}{}
			e := json.Unmarshal(msg, &t)
			if e != nil {
				continue
			}

			ret := global.Depth{
				Base:  sreq.Base,
				Quote: sreq.Quote,
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
		}
	}()

	return ch, nil
}

// SubLateTrade 查询交易详细数据
func (c *WSSClient) SubLateTrade(sreq global.TradeSymbol) (chan global.LateTrade, error) {
	conn, err := c.wsConnect()
	if err != nil {
		return nil, err
	}
	symbol := strings.ToLower(sreq.Base + sreq.Quote)

	topic := fmt.Sprintf("market.%s.trade.detail", symbol)
	req := struct {
		Topic string `json:"sub"`
		ID    string `json:"id"`
	}{topic, c.generateClientID()}

	err = conn.WriteJSON(req)
	if err != nil {
		c.Close()
		return nil, err
	}
	ch := make(chan global.LateTrade, 100)
	go func() {
		for {
			msg, err := c.readWSMessage(conn)
			if err != nil {
				log.Printf("SubLateTrade websocket 连接断开 %s\n", err.Error())
				go func() {
					for {
						time.Sleep(5 * time.Second)
						ch, err = c.SubLateTrade(sreq)
						if err == nil {
							log.Println("重新连接成功...")
							return
						}
					}
				}()
				return
			}
			if msg == nil {
				continue
			}
			// 解析数据
			rsp := struct {
				Tick struct {
					Data []struct {
						Price     float64 `json:"price"`
						Time      int64   `json:"ts"`
						Amount    float64 `json:"amount"`
						Direction string  `json:"direction"`
					} `json:"data"`
				} `json:"tick"`
			}{}
			err = json.Unmarshal(msg, &rsp)
			if err != nil {
				continue
			}

			for _, d := range rsp.Tick.Data {
				tm := time.Unix(d.Time/1000, 0)
				dt := tm.Format("2006-01-02 03:04:05 PM")
				lt := global.LateTrade{
					Base:      sreq.Base,
					Quote:     sreq.Quote,
					DateTime:  dt,
					Num:       d.Amount,
					Price:     d.Price,
					Dircetion: d.Direction,
					Total:     d.Price * d.Amount,
				}
				ch <- lt
			}
		}
	}()
	return ch, nil
}

// SubTicker ...
func (c *WSSClient) SubTicker(sreq global.TradeSymbol) (chan global.Ticker, error) {
	conn, err := c.wsConnect()
	if err != nil {
		return nil, err
	}
	symbol := strings.ToLower(sreq.Base + sreq.Quote)
	topic := fmt.Sprintf("market.%s.detail", symbol)
	req := struct {
		Topic string `json:"sub"`
		ID    string `json:"id"`
	}{topic, c.generateClientID()}

	err = conn.WriteJSON(req)
	if err != nil {
		c.Close()
		return nil, err
	}
	ch := make(chan global.Ticker, 100)

	go func() {
		for {
			msg, err := c.readWSMessage(conn)
			if err != nil {
				log.Printf("SubTicker websocket 连接断开 %s\n", err.Error())
				go func() {
					for {
						time.Sleep(5 * time.Second)
						ch, err = c.SubTicker(sreq)
						if err == nil {
							log.Println("重新连接成功...")
							return
						}
					}
				}()
				return
			}
			if msg == nil {
				continue
			}
			// 解析数据
			t := struct {
				Ticker struct {
					Amount float64 `json:"amount"`
					Open   float64 `json:"open"`
					Close  float64 `json:"close"`
					High   float64 `json:"high"`
					Low    float64 `json:"low"`
					Vol    float64 `json:"vol"`
				} `json:"tick"`
			}{}
			err = json.Unmarshal(msg, &t)
			if err != nil {
				continue
			}

			v := t.Ticker.Close - t.Ticker.Open
			ret := global.Ticker{
				Base:               sreq.Base,
				Quote:              sreq.Quote,
				PriceChange:        v,
				PriceChangePercent: v / t.Ticker.Open * 100,
				LastPrice:          t.Ticker.Close,
				HighPrice:          t.Ticker.High,
				LowPrice:           t.Ticker.Low,
				Volume:             t.Ticker.Vol,
			}
			ch <- ret
		}
	}()

	return ch, nil
}

func (c *WSSClient) wsConnect() (*websocket.Conn, error) {
	u := url.URL{Scheme: "wss", Host: *c.config.WSSHost, Path: "/ws"}
	log.Printf("huobi 连接 %s 中...\n", u.String())
	conn, _, err := c.config.WSSDialer.Dial(u.String(), nil)
	return conn, err
}

func (c *WSSClient) readWSMessage(conn *websocket.Conn) ([]byte, error) {
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
