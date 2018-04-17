package okex

import (
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

	return cfg
}
