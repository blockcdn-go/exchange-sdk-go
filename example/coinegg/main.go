package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/blockcdn-go/exchange-sdk-go/global"

	"github.com/blockcdn-go/exchange-sdk-go/clean"
	"github.com/blockcdn-go/exchange-sdk-go/coinegg"
	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/gorilla/websocket"
)

func main() {
	cfg := &config.Config{}

	sa1 := "s6nst-tupdj-e7p8q-dxdvh-p8bk7-bk2n4-7fau7"
	sa2 := "6Q3Y2-Ndhmx-(RsrH-/~YdM-)Ff1b-phKL5-ZRU;y"
	cfg.WithAPIKey(sa1)
	cfg.WithSecret(sa2)
	dialer := websocket.DefaultDialer
	u, _ := url.Parse("http://127.0.0.1:1080")
	dialer.Proxy = http.ProxyURL(u)
	cfg.WithWSSDialer(dialer)
	transport := clean.DefaultPooledTransport()
	transport.Proxy = http.ProxyURL(u)
	cfg.WithHTTPClient(&http.Client{Transport: transport})
	c := coinegg.NewClient(cfg)

	sm, err := c.GetAllSymbol()
	fmt.Printf("err %+v, %+v\n", err, sm)

	f, err := c.GetFund(global.FundReq{})
	fmt.Printf("err %+v, %+v\n", err, f)
}
