package huobi

import (
	"errors"
	"net/url"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/gorilla/websocket"
)

// WSSClient 是huobi sdk的调用客户端
type WSSClient struct {
	config config.Config
	conn   *websocket.Conn

	shouldQuit chan struct{}
	done       chan struct{}
}

// NewWSSClient 创建一个新的websocket客户端
func NewWSSClient(config *config.Config) *WSSClient {
	cfg := defaultConfig()
	if config != nil {
		cfg.MergeIn(config)
	}

	return &WSSClient{
		config:     *cfg,
		shouldQuit: make(chan struct{}),
		done:       make(chan struct{}),
	}
}

// QueryMarketKLine 查询市场K线图
func (c *WSSClient) QueryMarketKLine(symbol string, period string) (<-chan []byte, error) {
	err := c.connect()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (c *WSSClient) connect() error {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	u := url.URL{Scheme: "wss", Host: *c.config.WSSHost, Path: "/ws"}
	conn, _, err := c.config.WSSDialer.Dial(u.String(), nil)
	if err == nil {
		c.conn = conn
		return nil
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
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
