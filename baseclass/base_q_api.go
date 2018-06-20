package baseclass

import "github.com/blockcdn-go/exchange-sdk-go/global"

// GetKline 获取k线数据
func (c *Client) GetKline(req global.KlineReq) ([]global.Kline, error) {
	return c.Client.AicoinGetKline(c.Exchange, req)
}

// GetDepth 获取深度行情
func (c *Client) GetDepth(req global.TradeSymbol) (global.Depth, error) {
	return c.Client.AicoinGetDepth(c.Exchange, req)
}
