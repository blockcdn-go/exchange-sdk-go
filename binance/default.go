package binance

import (
	"github.com/blockcdn-go/exchange-sdk-go/clean"
	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/gorilla/websocket"
)

func defaultConfig() *config.Config {
	cfg := &config.Config{}
	cfg.WithWSSHost("stream.binance.com:9443")
	cfg.WithWSSDialer(websocket.DefaultDialer)
	cfg.WithHTTPClient(clean.DefaultPooledClient())
	cfg.WithUseSSL(true)
	return cfg
}
