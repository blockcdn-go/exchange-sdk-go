package weex

import "github.com/blockcdn-go/exchange-sdk-go/global"

// GetAllSymbol 交易市场详细行情接口
func (c *Client) GetAllSymbol() ([]global.TradeSymbol, error) {
	return []global.TradeSymbol{
		{Base: "BTC", Quote: "USD"},
		{Base: "BCH", Quote: "USD"},
		{Base: "LTC", Quote: "USD"},
		{Base: "ETH", Quote: "USD"},
		{Base: "ZEC", Quote: "USD"},
		{Base: "DASH", Quote: "USD"},
		{Base: "ETC", Quote: "USD"},
		{Base: "BCH", Quote: "BTC"},
		{Base: "LTC", Quote: "BTC"},
		{Base: "ETH", Quote: "BTC"},
		{Base: "ZEC", Quote: "BTC"},
		{Base: "DASH", Quote: "BTC"},
		{Base: "ETC", Quote: "BTC"},
	}, nil
}
