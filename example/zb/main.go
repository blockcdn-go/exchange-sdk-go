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

	sa1 := "23f7e16f-6145-49b2-b2d1-8e4faa7376de"
	sa2 := "ffaa6b9d-e38f-4bf4-a355-c85c3b9db511"
	cfg.WithAPIKey(sa1)
	cfg.WithSecret(sa2)
	dialer := websocket.DefaultDialer
	u, _ := url.Parse("http://127.0.0.1:1080")
	dialer.Proxy = http.ProxyURL(u)
	cfg.WithWSSDialer(dialer)
	transport := clean.DefaultPooledTransport()
	transport.Proxy = http.ProxyURL(u)
	cfg.WithHTTPClient(&http.Client{Transport: transport})
	c := zb.NewClient(cfg)

	sm, err := c.GetAllSymbol()
	fmt.Printf("err %+v, %+v\n", err, sm)

	dp, err := c.GetDepth(sm[0])
	fmt.Printf("err %+v, %+v\n", err, dp)

	kl, err := c.GetKline(global.KlineReq{
		Base:   sm[0].Base,
		Quote:  sm[0].Quote,
		Period: "1m",
	})
	fmt.Printf("err %+v, %+v\n", err, kl)

	f, err := c.GetFund(global.FundReq{})
	fmt.Printf("err %+v, %+v\n", err, f)

	i, err := c.InsertOrder(global.InsertReq{
		Base:  sm[0].Base,
		Quote: sm[0].Quote,
		Price: 1,
		Num:   1,
	})
	fmt.Printf("err %+v, %+v\n", err, i)

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
