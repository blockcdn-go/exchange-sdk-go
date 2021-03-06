package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/blockcdn-go/exchange-sdk-go/clean"
	"github.com/blockcdn-go/exchange-sdk-go/global"
	//	"net/http"
	//	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/blockcdn-go/exchange-sdk-go/huobi"
	"github.com/gorilla/websocket"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)

	dialer := websocket.DefaultDialer
	u, _ := url.Parse("http://127.0.0.1:1080")
	dialer.Proxy = http.ProxyURL(u)
	f, _ := os.Open("../cfg.json")
	js, _ := ioutil.ReadAll(f)
	cjs := struct {
		Huobi struct {
			APIKey string
			APISec string
		}
	}{}
	json.Unmarshal(js, &cjs)

	cfg := &config.Config{}
	cfg.WithAPIKey(cjs.Huobi.APIKey)
	cfg.WithSecret(cjs.Huobi.APISec)
	//c := huobi.NewClient(cfg)
	//r1, e1 := c.GetAllAccountID()
	//fmt.Println("GetAllAccountID: ", r1, e1)

	//	r2, e2 := c.BalanceInfo(false, 3270437)
	//	fmt.Println("BalanceInfo: ", r2, e2)
	/*
		req := huobi.InsertOrderReq{
			AccountID: "3270437",
			Amount:    "1",
			Price:     "",
			Source:    "api",
			Symbol:    "eosusdt",
			OrderType: "buy-market"}
		r3, e3 := c.InsertOrder(false, req)
		fmt.Println("InsertOrder: ", r3, e3)
	*/
	//	r3, e3 := c.GetOrders("eosusdt", "canceled")
	//	fmt.Println("GetOrders: ", r3, e3)

	//	time.Sleep(5 * time.Second)
	//	r4, e4 := c.GetOrderDetail(r3)
	//	fmt.Println("GetOrderDetail: ", r4, e4)

	//	r5 := c.CancelOrder(r3)
	//	fmt.Println("CancelOrder: ", r5)

	// r6, e6 := c.GetMatchDetail("3640838737")
	// fmt.Println("GetMatchDetail: ", r6, e6)

	cfg.WithWSSDialer(dialer)
	transport := clean.DefaultPooledTransport()
	transport.Proxy = http.ProxyURL(u)
	cfg.WithHTTPClient(&http.Client{Transport: transport})
	wss := huobi.NewClient(cfg)

	pair := global.TradeSymbol{Base: "btc", Quote: "usdt"}

	msgCh, err := wss.SubTicker(pair)
	if err != nil {
		log.Fatal("query error: ", err)
	}
	msgCh1, e1 := wss.SubDepth(pair)
	if e1 != nil {
		log.Fatal("query error: ", e1)
	}
	msgCh2, e2 := wss.SubLateTrade(pair)
	if e2 != nil {
		log.Fatal("query error: ", e2)
	}

	k, err := wss.GetKline(global.KlineReq{
		Base:   pair.Base,
		Quote:  pair.Quote,
		Period: "1m",
	})
	fmt.Printf("%+v, %+v\n", err, k)

	gd, e3 := wss.GetDepth(pair)
	fmt.Printf("%+v %+v\n", e3, gd)

	for {
		select {
		case <-interrupt:
			return
		case m := <-msgCh:
			fmt.Printf("Ticker: %+v\n", m)
		case m1 := <-msgCh1:
			fmt.Printf("Depth: %+v\n", m1)
		case m2 := <-msgCh2:
			fmt.Printf("LateTrade: %+v\n", m2)
		}
	}

}
