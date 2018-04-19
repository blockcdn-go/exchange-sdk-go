package main

import (
	"fmt"
	"log"
	//	"net/http"
	//	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/blockcdn-go/exchange-sdk-go/binance"
	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/gorilla/websocket"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)

	cfg := &config.Config{}
	dialer := websocket.DefaultDialer
	//	u, _ := url.Parse("http://127.0.0.1:1087")
	//	dialer.Proxy = http.ProxyURL(u)

	cfg.WithWSSDialer(dialer)
	c := binance.NewWSSClient(cfg)
	msgCh, err := c.KlineCandlestick("bnbbtc", "113231")
	if err != nil {
		log.Fatal("read error: ", err)
	}

	for {
		select {
		case <-interrupt:
			c.Close()
			return
		case m := <-msgCh:
			fmt.Println(string(m))
		}
	}
}
