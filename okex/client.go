package okex

import (
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSSClient 提供okex API调用的客户端
type WSSClient struct {
	config Config
	ticker *time.Ticker
	conn   *websocket.Conn

	shouldQuit chan struct{}
	retry      chan struct{}
	done       chan struct{}
	serverMu   sync.Mutex
}

// NewWSSClient 创建一个新的Websocket client
func NewWSSClient(config *Config) *WSSClient {
	cfg := DefaultConfig()
	if config != nil {
		cfg.MergeIn(config)
	}

	return &WSSClient{
		config:     *cfg,
		shouldQuit: make(chan struct{}),
		retry:      make(chan struct{}),
		done:       make(chan struct{}),
	}
}

// Query 负责订阅行情数据
func (c *WSSClient) Query() (<-chan string, error) {
	err := c.connect("/websocket")
	if err != nil {
		return nil, err
	}

	result := make(chan string)

	go c.ping()

	e := event{
		Event: "addChannel",
		Parameters: parameter{
			Base:    "okb",
			Binary:  "0",
			Product: "spot",
			Quote:   "btc",
			Type:    "ticker",
		},
	}
	err = c.conn.WriteJSON(e)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *WSSClient) ping() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.conn.WriteMessage(websocket.TextMessage, []byte("{'event':'ping'}"))
		case <-c.shouldQuit:
			c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}
	}
}

// Close 向服务端发起关闭操作
func (c *WSSClient) Close() {
	if c.conn == nil {
		return
	}

	close(c.shouldQuit)

	select {
	case <-c.done:
	case <-time.After(time.Second):
	}
}

func (c *WSSClient) connect(path string) error {
	c.serverMu.Lock()
	defer c.serverMu.Unlock()

	if c.conn != nil {
		c.conn = nil
	}

	u := url.URL{Scheme: "wss", Host: *c.config.WSSHost, Path: path}
	conn, _, err := c.config.WSSDialer.Dial(u.String(), nil)
	if err == nil {
		c.conn = conn
		return nil
	}

	for {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		select {
		case <-ticker.C:
			conn, _, err := c.config.WSSDialer.Dial(u.String(), nil)
			if err == nil {
				c.conn = conn
				return nil
			}
		case <-c.shouldQuit:
			return errors.New("Connection is closing")
		}
	}
}
