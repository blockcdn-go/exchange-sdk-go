package main

import (
	"fmt"
	//"log"

	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/blockcdn-go/exchange-sdk-go/gate"
)

func main() {
	cfg := &config.Config{}
	cfg.WithAPIKey("BBE2782F-69D1-445A-96E0-31D8DA35242E")
	cfg.WithSecret("00d37f5cb1a66f3329dca89260a0ee302e4ed9285f93caf6bd22a1d116aa4301")

	c := gate.NewClient(cfg)

	/*
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
	*/
	//////////////////////////////////

	b, e2 := c.BalanceInfo()
	fmt.Println("BalanceInfo", b, e2)

	//	r3, e3 := c.DepositAddr("bcdn")
	//	fmt.Println("DepositAddr ", r3, e3)

	//	r4, d4, e4 := c.DepositsWithdrawals()
	//	fmt.Println("DepositsWithdrawals ", r4, d4, e4)

	//	r5, e5 := c.InsertOrder("bcdn_usdt", 1, 10, 1)
	//	fmt.Println("InsertOrder ", r5, e5)

	//	r6, e6 := c.OrderStatusInfo("bcdn_usdt", "528502122")
	//	fmt.Println("OrderStatusInfo ", r6, e6)

	r7, e7 := c.HangingOrderInfo()
	fmt.Println("HangingOrderInfo ", r7, e7)

	//	r8, e8 := c.MatchInfo("bcdn_usdt", "")
	//	fmt.Println("MatchInfo ", r8, e8)

	fmt.Println("CancelOrder ", c.CancelOrder("bcdn_usdt", "528514518"))
}
