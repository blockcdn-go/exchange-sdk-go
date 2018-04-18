package huobi

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WSSClient 是huobi sdk的调用客户端
type WSSClient struct {
	config config.Config
	conns  map[string]*websocket.Conn

	closed  bool
	closeMu sync.Mutex

	shouldQuit chan struct{}
	done       chan struct{}
	retry      chan string
}

// NewWSSClient 创建一个新的websocket客户端
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

// Close 发起关闭操作
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

func (c *WSSClient) start(req interface{}, cid string, msgCh chan<- []byte) {
	go c.query(cid, msgCh)

	for {
		select {
		case cid := <-c.retry:
			delete(c.conns, cid)
			c.reconnect(req, msgCh)
		case <-c.shouldQuit:
			c.shutdown()
			return
		}
	}
}

func (c *WSSClient) reconnect(req interface{}, msgCh chan<- []byte) {
	cid, conn, err := c.connect()
	if err != nil {
		return
	}

	err = conn.WriteJSON(req)
	if err != nil {
		c.closeMu.Lock()
		defer c.closeMu.Unlock()

		if !c.closed {
			c.retry <- cid
		}
		return
	}

	go c.query(cid, msgCh)
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

		buf := bytes.NewBuffer(msg)
		gz, err := gzip.NewReader(buf)
		if err != nil {
			continue
		}
		message, _ := ioutil.ReadAll(gz)

		if strings.Contains(string(message), "ping") {
			c.pong(cid, conn, message)
			continue
		}

		msgCh <- message
	}
}

func (c *WSSClient) pong(cid string, conn *websocket.Conn, msg []byte) {
	var ping struct {
		Ping int64 `json:"ping"`
	}

	err := json.Unmarshal(msg, &ping)
	if err != nil {
		c.closeMu.Lock()
		defer c.closeMu.Unlock()

		if !c.closed {
			c.retry <- cid
		}

		return
	}

	pong := struct {
		Pong int64 `json:"pong"`
	}{ping.Ping}
	conn.WriteJSON(pong)
}

func (c *WSSClient) connect() (string, *websocket.Conn, error) {
	u := url.URL{Scheme: "wss", Host: *c.config.WSSHost, Path: "/ws"}
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

func (c *WSSClient) generateClientID() string {
	now := time.Now().UnixNano()
	return strconv.FormatInt(now, 10)
}
