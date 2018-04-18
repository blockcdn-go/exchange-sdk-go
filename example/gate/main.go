package main

import (
	"fmt"
	"log"

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
	ticker, e := c.TickerInfo("btc","usdt")
	if  e == nil {
		fmt.Printf("ticker:%v\n",ticker)
	}else{
		fmt.Print(e)
	}
	depth, e1 := c.DepthInfo("btc","usdt")
	if  e1 == nil {
		fmt.Printf("depth:%v\n",depth)
	}else{
		fmt.Print(e1)
	}
	b, e2 := c.BalanceInfo()
	if  e2 == nil {
		fmt.Printf("BalanceInfo:%v\n",b)
	}else{
		fmt.Print(e2)
	}

	//
	addr, e3 := c.DepositAddr("btc")
	if  e3 == nil {
		fmt.Printf("DepositAddr:%v\n",addr)
	}else{
		fmt.Println(e3)
	}

	d, w, e4 := c.DepositsWithdrawals()
	if  e4 == nil {
		fmt.Printf("DepositsWithdrawals:%v, \n %v\n",d, w)
	}else{
		fmt.Println(e4)
	}
}
