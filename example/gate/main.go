package main

import (
	"fmt"
	"log"
	//"log"

	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/blockcdn-go/exchange-sdk-go/gate"
)

func main() {
	cfg := &config.Config{}
	cfg.WithAPIKey("AB8DEF71-78E8-4C07-B873-36F4ACE8A0E3")
	cfg.WithSecret("dd6223e511a2b7592c4c111358fcc0f7e86cf20df83b4264d2d20dee439035bc")

	c := gate.NewClient(cfg)

	resp, err := c.MarketList()
	if err != nil {
		log.Fatal("error: ", err)
	}

	for _, v := range resp {
		fmt.Printf("%v\n", v)
	}
	ticker, e := c.TickerInfo("btc", "usdt")
	fmt.Print("ticker ", ticker, e)

	depth, e1 := c.DepthInfo("btc", "usdt")
	fmt.Print("DepthInfo", depth, e1)

	//////////////////////////////////

	b, e2 := c.BalanceInfo()
	fmt.Print("BalanceInfo", b, e2)

	r3, e3 := c.DepositAddr("btc")
	fmt.Println("DepositAddr ", r3, e3)

	r4, d4, e4 := c.DepositsWithdrawals()
	fmt.Println("DepositsWithdrawals ", r4, d4, e4)

	r5, e5 := c.InsertOrder("btc_usdt", 0, 1000, 1)
	fmt.Println("InsertOrder ", r5, e5)

	fmt.Println("CancelOrder ", c.CancelOrder("btc_usdt", "00000"))
}
