package main

import (
	"fmt"

	"github.com/blockcdn-go/exchange-sdk-go/global"

	"github.com/blockcdn-go/exchange-sdk-go/bitstamp"
	"github.com/blockcdn-go/exchange-sdk-go/config"
)

func main() {
	cfg := &config.Config{}

	sa1 := "s6nst-tupdj-e7p8q-dxdvh-p8bk7-bk2n4-7fau7"
	sa2 := "6Q3Y2-Ndhmx-(RsrH-/~YdM-)Ff1b-phKL5-ZRU;y"
	cfg.WithAPIKey(sa1)
	cfg.WithSecret(sa2)

	c := bitstamp.NewClient(cfg)

	sm, err := c.GetAllSymbol()
	fmt.Printf("err %+v, %+v\n", err, sm)

	kl, err := c.GetKline(global.KlineReq{Base: "btc", Quote: "usd", Period: "1m"})
	fmt.Printf("err %+v, %+v\n", err, kl)

	dp, err := c.GetDepth(global.TradeSymbol{Base: "btc", Quote: "usd"})
	fmt.Printf("err %+v, %+v\n", err, dp)

	tk, err := c.AicoinGetTicker("bitstamp", global.TradeSymbol{Base: "btc", Quote: "usd"})
	fmt.Printf("err %+v, %+v\n", err, tk)
	// f, err := c.GetFund(global.FundReq{})
	// fmt.Printf("err %+v, %+v\n", err, f)
}
