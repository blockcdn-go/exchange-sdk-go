package baseclass

import (
	"sync"

	"github.com/blockcdn-go/exchange-sdk-go/aicoin"
	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/blockcdn-go/exchange-sdk-go/global"
)

// Client 提供 API的调用客户端
type Client struct {
	aicoin.Client
	Exchange string // 子类设置
	mutex    sync.Mutex
	lastt    map[global.TradeSymbol][]global.LateTrade
}

// Constructor 创建一个新的client
func (c *Client) Constructor(config *config.Config) {
	c.Client.Constructor(config)
	c.lastt = make(map[global.TradeSymbol][]global.LateTrade)
}
