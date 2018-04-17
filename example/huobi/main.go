package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
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
	u, _ := url.Parse("http://127.0.0.1:8118")
	dialer.Proxy = http.ProxyURL(u)

	cfg := &config.Config{}
	cfg.WithWSSDialer(dialer)
	wss := huobi.NewWSSClient(cfg)
	msgCh, err := wss.QueryMarketKLine("ethbtc", huobi.Period1Min)
	if err != nil {
		log.Fatal("query error: ", err)
	}

	for {
		select {
		case <-interrupt:
			wss.Close()
			return
		case m := <-msgCh:
			fmt.Println(string(m))
		}
	}
}
