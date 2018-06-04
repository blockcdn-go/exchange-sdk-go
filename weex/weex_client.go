package weex

import (
	"github.com/blockcdn-go/exchange-sdk-go/config"
)

// Client 提供weex API的调用客户端
type Client struct {
	config config.Config
}

// NewClient 创建一个新的client
func NewClient(config *config.Config) *Client {
	cfg := defaultConfig()
	if config != nil {
		cfg.MergeIn(config)
	}

	return &Client{config: *cfg}
}
