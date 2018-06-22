package main

import (
	"fmt"

	"github.com/blockcdn-go/exchange-sdk-go/coinex"
	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/blockcdn-go/exchange-sdk-go/global"
)

func main() {
	cfg := &config.Config{}

	sa1 := "EA2190E24A72466584349227A503D65F"
	sa2 := "00d37f5cb1a66f3329dca89260a0ee302e4ed9285f93caf6bd22a1d116aa4301"
	cfg.WithAPIKey(sa1)
	cfg.WithSecret(sa2)

	c := coinex.NewClient(cfg)

	sm, err := c.GetAllSymbol()
	fmt.Printf("err %+v, %+v\n", err, sm)

	kl, err := c.GetKline(global.KlineReq{Base: "btc", Quote: "usdt", Period: "1m"})
	fmt.Printf("err %+v, %+v\n", err, kl)

	dp, err := c.GetDepth(global.TradeSymbol{Base: "btc", Quote: "usdt"})
	fmt.Printf("err %+v, %+v\n", err, dp)

	f, err := c.GetFund(global.FundReq{})
	fmt.Printf("err %+v, %+v\n", err, f)
}
