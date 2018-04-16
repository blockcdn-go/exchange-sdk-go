package okex

import (
	"bytes"
	"compress/flate"
	"errors"
	"io/ioutil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSSClient 提供okex API调用的客户端
type WSSClient struct {
	config Config
	ticker *time.Ticker
	conn   *websocket.Conn

	closed  bool
	closeMu sync.Mutex

	shouldQuit chan struct{}
	retry      chan struct{}
	done       chan struct{}
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

// QuerySpot 负责订阅现货行情数据
func (c *WSSClient) QuerySpot() (<-chan []byte, error) {
	err := c.connect("/websocket")
	if err != nil {
		return nil, err
	}

	result := make(chan []byte)
	go c.start("/websocket", result)

	err = c.subscribeSpot()
	if err != nil {
		c.Close()
		return nil, err
	}

	return result, nil
}

func (c *WSSClient) subscribeSpot() error {
	for _, v := range events {
		err := c.conn.WriteJSON(v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *WSSClient) start(path string, msgCh chan<- []byte) {
	go c.query(msgCh)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.conn.WriteMessage(websocket.TextMessage, []byte("{'event':'ping'}"))
		case <-c.retry:
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

	c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	c.conn.Close()
	close(c.done)
}

func (c *WSSClient) reconnect(path string, msgCh chan<- []byte) {
	err := c.connect(path)
	if err != nil {
		return
	}

	go c.query(msgCh)
	err = c.subscribeSpot()
	if err != nil {
		c.closeMu.Lock()
		defer c.closeMu.Unlock()

		if !c.closed {
			c.retry <- struct{}{}
		}
	}
}

func (c *WSSClient) query(msgCh chan<- []byte) {
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			c.closeMu.Lock()
			defer c.closeMu.Unlock()

			if !c.closed {
				c.retry <- struct{}{}
			}
			return
		}

		if strings.Contains(string(msg), "pong") {
			msgCh <- msg
			continue
		}

		buf := bytes.NewBuffer(msg)
		z := flate.NewReader(buf)
		m, _ := ioutil.ReadAll(z)
		msgCh <- m
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
	if c.conn != nil {
		c.conn.Close()
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
