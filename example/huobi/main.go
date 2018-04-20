package main

import (
	"fmt"
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

	cfg := &config.Config{}
	cfg.WithAPIKey("c4299337-78f7c704-96e370ad-9f8fa")
	cfg.WithSecret("dc442c61-f6cb0dac-14091e27-4b955")
	c := huobi.NewClient(cfg)
	r1, e1 := c.GetAllAccountID()
	fmt.Println("GetAllAccountID: ", r1, e1)

	r2, e2 := c.BalanceInfo(true, 3270437)
	fmt.Println("BalanceInfo: ", r2, e2)

	req := huobi.InsertOrderReq{"0", "1", "2", "3", "4", "5"}
	r3, e3 := c.InsertOrder(false, req)
	fmt.Println("InsertOrder: ", r3, e3)
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
