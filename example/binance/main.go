package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"time"

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

	// interrupt := make(chan os.Signal, 1)
	// signal.Notify(interrupt, os.Interrupt)

	// kech, err := b.TradeWebsocket("BTCUSDT")
	// if err != nil {
	// 	panic(err)
	// }
	// depth, _ := b.DepthWebsocket("BTCUSDT")

	// tk, _ := b.TickerWebsocket("BTCUSDT")
	// go func() {
	// 	for {
	// 		select {
	// 		case ke := <-kech:
	// 			fmt.Printf("%+v\n", ke)
	// 		case d := <-depth:
	// 			fmt.Printf("%+v\n", d)
	// 		case t := <-tk:
	// 			fmt.Printf("%+v\n", t)
	// 		}
	// 	}
	// }()

	// fmt.Println("waiting for interrupt")
	// <-interrupt
	// fmt.Println("canceling context")

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
	r0, e0 := b.Time()
	fmt.Println("time:", r0, e0)
	r1, e1 := b.AllOrders(binance.AllOrdersRequest{
		Symbol:  "FUNETH",
		OrderID: 12222764,
	})
	fmt.Println("allorders:", r1, e1)

	newo := binance.NewOrderRequest{
		Symbol:      "FUNETH",
		Quantity:    200,
		Price:       0.00006179,
		Side:        binance.SideBuy,
		TimeInForce: binance.GTC,
		Type:        binance.TypeLimit,
		Timestamp:   time.Now(),
	}
	//	err := b.NewOrderTest(newo)
	//	fmt.Println("NewOrderTest:", err)

	newOrder, err := b.NewOrder(newo)

	fmt.Println("NewOrder:", newOrder, err)
	if err != nil {
		panic("...")
	}
	orderid := int64(12222764)

	// pass
	res2, err := b.QueryOrder(binance.QueryOrderRequest{
		Symbol:  "FUNETH",
		OrderID: orderid,
		//RecvWindow: 5 * time.Second,
		Timestamp: time.Now(),
	})
	fmt.Println("QueryOrder:", res2, err)

	// pass
	res4, err := b.OpenOrders(binance.OpenOrdersRequest{
		Symbol:     "FUNETH",
		RecvWindow: 5 * time.Second,
		Timestamp:  time.Now(),
	})
	fmt.Println("OpenOrders:", res4, err)

	// pass
	res3, err := b.CancelOrder(binance.CancelOrderRequest{
		Symbol:    "FUNETH",
		OrderID:   orderid,
		Timestamp: time.Now(),
	})
	fmt.Println("cancel order:", res3, err)

	// pass
	res6, err := b.Account(binance.AccountRequest{
		RecvWindow: 5 * time.Second,
		Timestamp:  time.Now(),
	})
	fmt.Println("Account:", res6, err)

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
