package huobi

import (
	"github.com/blockcdn-go/exchange-sdk-go/clean"
	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/gorilla/websocket"
)

func defaultConfig() *config.Config {
	cfg := &config.Config{}
	cfg.WithWSSHost("api.hadax.com")
	cfg.WithRESTHost("api.hadax.com")
	cfg.WithHTTPClient(clean.DefaultPooledClient())
	cfg.WithWSSDialer(websocket.DefaultDialer)
	cfg.WithUseSSL(true)
	return cfg
}
