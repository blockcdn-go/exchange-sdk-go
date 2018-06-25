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
	sa2 := "7629CE49CA854B54B945EEE0E77F3A6806C0F388C0536B38"
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

	i, err := c.InsertOrder(global.InsertReq{
		Base:  "btc",
		Quote: "usdt",
		Price: 5000,
		Num:   1,
	})
	fmt.Printf("err %+v, %+v\n", err, i)

	err = c.CancelOrder(global.CancelReq{
		Base:    "btc",
		Quote:   "usdt",
		OrderNo: "111",
	})
	fmt.Printf("err %+v\n", err)

	s, err := c.OrderStatus(global.StatusReq{
		Base:    "btc",
		Quote:   "usdt",
		OrderNo: "111",
	})
	fmt.Printf("err %+v, %+v\n", err, s)
}
