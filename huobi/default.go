package huobi

import (
	"net/http"
	"net/url"

	"github.com/blockcdn-go/exchange-sdk-go/clean"
	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/gorilla/websocket"
)

func defaultConfig() *config.Config {
	cfg := &config.Config{}
	cfg.WithWSSHost("api.huobipro.com")
	cfg.WithRESTHost("api.huobipro.com")
	transport := clean.DefaultPooledTransport()
	u, _ := url.Parse("http://127.0.0.1:1080")
	transport.Proxy = http.ProxyURL(u)
	cfg.WithHTTPClient(&http.Client{Transport: transport})
	//cfg.WithHTTPClient(clean.DefaultPooledClient())
	cfg.WithWSSDialer(websocket.DefaultDialer)
	cfg.WithUseSSL(true)
	return cfg
}
