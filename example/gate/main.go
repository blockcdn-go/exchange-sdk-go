package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	//"log"

	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/blockcdn-go/exchange-sdk-go/gate"
)

func main() {
	f, _ := os.Open("../cfg.json")
	js, _ := ioutil.ReadAll(f)
	cjs := struct {
		Gate struct {
			APIKey string
			APISec string
		}
	}{}
	json.Unmarshal(js, &cjs)

	cfg := &config.Config{}
	cfg.WithAPIKey(cjs.Gate.APIKey)
	cfg.WithSecret(cjs.Gate.APISec)

	c := gate.NewClient(cfg)

	resp, err := c.GetAllSymbol()
	if err != nil {
		log.Fatal("error: ", err)
	}

	for _, v := range resp {
		fmt.Printf("%v\n", v)
	}
	ticker, e := c.TickerInfo("btc", "usdt")
	fmt.Print("ticker ", ticker, e)

	depth, e1 := c.DepthInfo("bcdn", "usdt")
	fmt.Print("DepthInfo", depth, e1)

	tch, err := c.SubTicker(resp[0])
	dch, err := c.SubDepth(resp[0])
	lch, err := c.SubLateTrade(resp[0])

	for {
		select {
		case tk := <-tch:
			fmt.Printf("ticker %+v\n", tk)
		case dp := <-dch:
			fmt.Printf("depth %+v\n", dp)
		case lt := <-lch:
			fmt.Printf("latetrade %+v\n", lt)
		}
	}
	//////////////////////////////////

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

}
