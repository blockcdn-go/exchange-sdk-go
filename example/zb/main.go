package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/blockcdn-go/exchange-sdk-go/clean"
	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/blockcdn-go/exchange-sdk-go/global"
	"github.com/blockcdn-go/exchange-sdk-go/zb"
	"github.com/gorilla/websocket"
)

func main() {
	cfg := &config.Config{}
	cfg.WithAPIKey("ea72f9d6-80a1-4bf7-8dfe-bd000c11e32f")
	cfg.WithSecret("D54DD52FBEAA4EF5B98598B62BAC70A0DBB11BC716E84E7D70")
	dialer := websocket.DefaultDialer
	u, _ := url.Parse("http://127.0.0.1:1080")
	dialer.Proxy = http.ProxyURL(u)
	cfg.WithWSSDialer(dialer)
	transport := clean.DefaultPooledTransport()
	transport.Proxy = http.ProxyURL(u)
	cfg.WithHTTPClient(&http.Client{Transport: transport})
	c := zb.NewClient(cfg)
	s := global.TradeSymbol{Base: "btc", Quote: "usdt"}
	ctk, err := c.SubTicker(s)
	fmt.Printf("err %+v\n", err)

	cdp, err := c.SubDepth(s)
	fmt.Printf("err %+v\n", err)

	clt, err := c.SubLateTrade(s)
	fmt.Printf("err %+v\n", err)

	for {
		select {
		case tk := <-ctk:
			fmt.Printf("ticker %+v\n", tk)
		case dp := <-cdp:
			fmt.Printf("depth %+v\n", dp)
		case lt := <-clt:
			fmt.Printf("trade %+v\n", lt)
		}
	}
}
