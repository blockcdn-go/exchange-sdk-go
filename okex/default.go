package okex

import (
	"net/http"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/clean"
	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/gorilla/websocket"
)

func defaultConfig() *config.Config {
	cfg := &config.Config{}

	cfg.WithWSSHost("okexcomreal.bafang.com:10441")
	cfg.WithHTTPClient(clean.DefaultPooledClient())
	cfg.WithWSSDialer(websocket.DefaultDialer)
	cfg.WithUseSSL(true)
	cfg.WithPingDuration(60 * time.Second)

	//
	cfg.WithRESTHost("www.okex.cn")
	cfg.WithUseSSL(true)
	transport := clean.DefaultPooledTransport()
	cfg.WithHTTPClient(&http.Client{Transport: transport})
	return cfg
}
