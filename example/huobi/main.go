package main

import (
	"fmt"
	"log"
	//	"net/http"
	//	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"

	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/blockcdn-go/exchange-sdk-go/huobi"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)

	dialer := websocket.DefaultDialer
	//	u, _ := url.Parse("http://127.0.0.1:8118")
	//	dialer.Proxy = http.ProxyURL(u)

	cfg := &config.Config{}
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
}
