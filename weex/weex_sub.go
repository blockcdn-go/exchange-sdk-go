package weex

import "github.com/blockcdn-go/exchange-sdk-go/global"

// SubTicker ...
func (c *Client) SubTicker(sreq global.TradeSymbol) (chan global.Ticker, error) {
	ch := make(chan global.Ticker, 100)

	return ch, nil
}
