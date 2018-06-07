package main

import (
	"fmt"

	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/blockcdn-go/exchange-sdk-go/global"

	"github.com/blockcdn-go/exchange-sdk-go/weex"
)

func main() {
	cfg := &config.Config{}
	cfg.WithAPIKey("ea72f9d6-80a1-4bf7-8dfe-bd000c11e32f")
	cfg.WithSecret("D54DD52FBEAA4EF5B98598B62BAC70A0DBB11BC716E84E7D70")
	c := weex.NewClient(cfg)
	s, err := c.GetAllSymbol()
	fmt.Printf("%+v, %+v\n", err, s)

	k, err := c.GetKline(global.KlineReq{
		Base:   s[0].Base,
		Quote:  s[0].Quote,
		Period: "1m",
	})
	fmt.Printf("%+v, %+v\n", err, k)

	f, err := c.GetFund(global.FundReq{})
	fmt.Printf("%+v, %+v\n", err, f)

	i, err := c.InsertOrder(global.InsertReq{
		APIKey: "ea72f9d6-80a1-4bf7-8dfe-bd000c11e32f",
		Base:   s[0].Base,
		Quote:  s[0].Quote,
		Price:  100,
		Num:    100,
	})
	fmt.Printf("%+v, %+v\n", err, i)

	o, err := c.OrderStatus(global.StatusReq{
		APIKey: "ea72f9d6-80a1-4bf7-8dfe-bd000c11e32f",
		Base:   s[0].Base,
		Quote:  s[0].Quote,
	})
	fmt.Printf("%+v, %+v\n", err, o)

	tch, err := c.SubTicker(s[0])
	dch, err := c.SubDepth(s[0])
	lch, err := c.SubLateTrade(s[0])

	for {
		select {
		case tk := <-tch:
			fmt.Printf("notify %+v\n", tk)
			break
		case td := <-dch:
			fmt.Printf("notify %+v\n", td)
			break
		case lt := <-lch:
			fmt.Printf("notify %+v\n", lt)
			break
		}
	}
}
