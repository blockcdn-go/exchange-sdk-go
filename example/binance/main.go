package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/global"

	"github.com/blockcdn-go/exchange-sdk-go/binance"
)

func main() {

	f, _ := os.Open("../cfg.json")
	js, _ := ioutil.ReadAll(f)
	cjs := struct {
		Binance struct {
			APIKey string
			APISec string
		}
	}{}
	json.Unmarshal(js, &cjs)
	pxy, _ := url.Parse("http://127.0.0.1:1080")

	//ctx, _ := context.WithCancel(context.Background())
	// use second return value for cancelling request
	b := binance.NewAPIService(
		nil,
		"https://www.binance.com",
		cjs.Binance.APIKey,
		cjs.Binance.APISec,
		pxy)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	ts := global.TradeSymbol{Base: "btc", Quote: "usdt"}
	kech, err := b.SubLateTrade(ts)
	fmt.Printf("%+v, %+v\n", kech, err)
	depth, err := b.SubDepth(ts)
	fmt.Printf("%+v, %+v\n", depth, err)
	tk, err := b.SubTicker(ts)
	fmt.Printf("%+v, %+v\n", tk, err)
	go func() {
		for {
			select {
			case ke := <-kech:
				fmt.Printf("SubLateTrade %+v\n", ke)
			case d := <-depth:
				fmt.Printf("SubDepth %+v\n", d)
			case t := <-tk:
				fmt.Printf("SubTicker %+v\n", t)
			}
		}
	}()

	fmt.Println("waiting for interrupt")
	<-interrupt
	fmt.Println("canceling context")

	// fmt.Println("exit")
	// return

	// kl, err := b.Klines(binance.KlinesRequest{
	// 	Symbol:   "BNBETH",
	// 	Interval: binance.Hour,
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("%#v\n", kl)

	// pass
	res6, err := b.GetFund(global.FundReq{})
	fmt.Println("Account:", res6, err)

	// pass
	k1, ke1 := b.GetKline(global.KlineReq{
		Base:   ts.Base,
		Quote:  ts.Quote,
		Period: "1m",
		Count:  500,
	})
	fmt.Println("Klines:", k1, ke1)

	r0, e0 := b.Time()
	fmt.Println("time:", r0, e0)
	r1, e1 := b.AllOrders(binance.AllOrdersRequest{
		Symbol:  "FUNETH",
		OrderID: 12222764,
	})
	fmt.Println("allorders:", r1, e1)

	newOrder, err := b.InsertOrder(global.InsertReq{
		Base:  ts.Base,
		Quote: ts.Quote,
	})

	fmt.Println("NewOrder:", newOrder, err)
	if err != nil {
		panic("...")
	}

	// pass
	res2, err := b.OrderStatus(global.StatusReq{})
	fmt.Println("QueryOrder:", res2, err)

	// pass
	res4, err := b.OpenOrders(binance.OpenOrdersRequest{
		Symbol:     "FUNETH",
		RecvWindow: 5 * time.Second,
		Timestamp:  time.Now(),
	})
	fmt.Println("OpenOrders:", res4, err)

	// pass
	err = b.CancelOrder(global.CancelReq{})
	fmt.Println("cancel order:", err)

	res7, err := b.MyTrades(binance.MyTradesRequest{
		Symbol:     "BNBETH",
		RecvWindow: 5 * time.Second,
		Timestamp:  time.Now(),
	})
	fmt.Println("MyTrades:", res7, err)

	res9, err := b.DepositHistory(binance.HistoryRequest{
		Timestamp:  time.Now(),
		RecvWindow: 5 * time.Second,
	})
	fmt.Println("DepositHistory:", res9, err)

	res8, err := b.WithdrawHistory(binance.HistoryRequest{
		Timestamp:  time.Now(),
		RecvWindow: 5 * time.Second,
	})
	fmt.Println("WithdrawHistory:", res8, err)

	ds, err := b.StartUserDataStream()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", ds)

	err = b.KeepAliveUserDataStream(ds)
	if err != nil {
		panic(err)
	}

	err = b.CloseUserDataStream(ds)
	if err != nil {
		panic(err)
	}
}
