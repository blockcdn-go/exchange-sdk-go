package binance

import (
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WSSClient 是binance websocket client
type WSSClient struct {
	config config.Config
	conns  map[string]*websocket.Conn

	closed  bool
	closeMu sync.Mutex

	shouldQuit chan struct{}
	done       chan struct{}
	retry      chan string
}

// NewWSSClient 创建一个新的binance websocket client
func NewWSSClient(config *config.Config) *WSSClient {
	cfg := defaultConfig()
	if config != nil {
		cfg.MergeIn(config)
	}

	return &WSSClient{
		config:     *cfg,
		conns:      make(map[string]*websocket.Conn),
		shouldQuit: make(chan struct{}),
		done:       make(chan struct{}),
		retry:      make(chan string),
	}
}

// AggregateTrade Aggregate Trade Streams
func (c *WSSClient) AggregateTrade(symbol string) (<-chan []byte, error) {
	p := fmt.Sprintf("/ws/%s@aggTrade", symbol)
	cid, _, err := c.connect(p)
	if err != nil {
		return nil, err
	}

	result := make(chan []byte)
	go c.start(cid, p, result)
	return result, nil
}

// Trade Streams
func (c *WSSClient) Trade(symbol string) (<-chan []byte, error) {
	p := fmt.Sprintf("/ws/%s@trade", symbol)
	cid, _, err := c.connect(p)
	if err != nil {
		return nil, err
	}

	result := make(chan []byte)
	go c.start(cid, p, result)
	return result, nil
}

// KlineCandlestick Kline/Candlestick Streams
func (c *WSSClient) KlineCandlestick(symbol string, period string) (<-chan []byte, error) {
	p := fmt.Sprintf("/ws/%s@kline_%s", symbol, period)
	cid, _, err := c.connect(p)
	if err != nil {
		return nil, err
	}

	result := make(chan []byte)
	go c.start(cid, p, result)
	return result, nil
}

// IndividualSymbolTicker Individual Symbol Ticker Streams
func (c *WSSClient) IndividualSymbolTicker(symbol string) (<-chan []byte, error) {
	p := fmt.Sprintf("/ws/%s@ticker", symbol)
	cid, _, err := c.connect(p)
	if err != nil {
		return nil, err
	}

	result := make(chan []byte)
	go c.start(cid, p, result)
	return result, nil
}

// AllMarketTickers All Market Tickers Stream
func (c *WSSClient) AllMarketTickers() (<-chan []byte, error) {
	cid, _, err := c.connect("/ws/!ticker@arr")
	if err != nil {
		return nil, err
	}

	result := make(chan []byte)
	go c.start(cid, "/ws/!ticker@arr", result)
	return result, nil
}

// PartialBookDepth Partial Book Depth Streams
func (c *WSSClient) PartialBookDepth(symbol string, level int) (<-chan []byte, error) {
	p := fmt.Sprintf("/ws/%s@depth%d", symbol, level)
	cid, _, err := c.connect(p)
	if err != nil {
		return nil, err
	}

	result := make(chan []byte)
	go c.start(cid, p, result)
	return result, nil
}

// DiffDepth Diff. Depth Stream
func (c *WSSClient) DiffDepth(symbol string) (<-chan []byte, error) {
	p := fmt.Sprintf("/ws/%s@depth", symbol)
	cid, _, err := c.connect(p)
	if err != nil {
		return nil, err
	}

	result := make(chan []byte)
	go c.start(cid, p, result)
	return result, nil
}

// Close 关闭websocket
func (c *WSSClient) Close() {
	if c.conns == nil || len(c.conns) == 0 {
		return
	}

	close(c.shouldQuit)

	select {
	case <-c.done:
	case <-time.After(time.Second):
	}
}

func (c *WSSClient) start(cid, path string, msgCh chan<- []byte) {
	go c.query(cid, msgCh)

	for {
		select {
		case cid := <-c.retry:
			delete(c.conns, cid)
			c.reconnect(path, msgCh)
		case <-c.shouldQuit:
			c.shutdown()
			return
		}
	}
}

func (c *WSSClient) shutdown() {
	c.closeMu.Lock()
	c.closed = true
	c.closeMu.Unlock()

	for _, conn := range c.conns {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		conn.Close()
	}

	close(c.done)
}

func (c *WSSClient) reconnect(path string, msgCh chan<- []byte) {
	cid, _, err := c.connect(path)
	if err != nil {
		return
	}

	go c.query(cid, msgCh)
}

func (c *WSSClient) query(cid string, msgCh chan<- []byte) {
	conn, ok := c.conns[cid]
	if !ok {
		return
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			c.closeMu.Lock()
			defer c.closeMu.Unlock()

			if !c.closed {
				c.retry <- cid
			}

			return
		}

		msgCh <- msg
	}
}

func (c *WSSClient) connect(path string) (string, *websocket.Conn, error) {
	u := url.URL{Scheme: "wss", Host: *c.config.WSSHost, Path: path}
	conn, _, err := c.config.WSSDialer.Dial(u.String(), nil)
	if err == nil {
		u := uuid.New().String()
		c.conns[u] = conn
		return u, conn, nil
	}

	if err == websocket.ErrBadHandshake {
		return "", nil, err
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			conn, _, err := c.config.WSSDialer.Dial(u.String(), nil)
			if err == nil {
				u := uuid.New().String()
				c.conns[u] = conn
				return u, conn, nil
			}

		case <-c.shouldQuit:
			return "", nil, errors.New("Connection is closing")
		}
	}
}
