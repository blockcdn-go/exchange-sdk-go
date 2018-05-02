package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	//	"net/http"
	//	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/blockcdn-go/exchange-sdk-go/huobi"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)

	//dialer := websocket.DefaultDialer
	//	u, _ := url.Parse("http://127.0.0.1:8118")
	//	dialer.Proxy = http.ProxyURL(u)
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
	c := huobi.NewClient(cfg)
	r1, e1 := c.GetAllAccountID()
	fmt.Println("GetAllAccountID: ", r1, e1)

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

	r6, e6 := c.GetMatchDetail("3640838737")
	fmt.Println("GetMatchDetail: ", r6, e6)
	/*
		cfg.WithWSSDialer(dialer)
		wss := huobi.NewWSSClient(cfg)
		msgCh, err := wss.SubMarketKLine("lxtbtc", huobi.Period1Min)
		if err != nil {
			log.Fatal("query error: ", err)
		}
		msgCh1, e1 := wss.SubMarketDepth("lxtbtc", "step5")
		if e1 != nil {
			log.Fatal("query error: ", e1)
		}
		msgCh2, e2 := wss.SubTradeDetail("lxtbtc")
		if e2 != nil {
			log.Fatal("query error: ", e2)
		}
		msgCh3, e3 := wss.SubMarketDetail("lxtbtc")
		if e3 != nil {
			log.Fatal("query error: ", e3)
		}
		for {
			select {
			case <-interrupt:
				wss.Close()
				return
			case m := <-msgCh:
				fmt.Println("SubMarketKLine: ", string(m))
			case m1 := <-msgCh1:
				fmt.Println("SubMarketDepth: ", string(m1))
			case m2 := <-msgCh2:
				fmt.Println("SubTradeDetail: ", string(m2))
			case m3 := <-msgCh3:
				fmt.Println("SubMarketDetail: ", string(m3))
			}
		}
	*/
}
