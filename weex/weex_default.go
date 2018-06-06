package weex

import (
	"net/http"

	"github.com/blockcdn-go/exchange-sdk-go/clean"
	"github.com/blockcdn-go/exchange-sdk-go/config"
)

func defaultConfig() *config.Config {
	cfg := &config.Config{}

	cfg.WithRESTHost("api.weex.com")
	cfg.WithSecret("")
	cfg.WithAPIKey("")
	transport := clean.DefaultPooledTransport()
	// u, _ := url.Parse("http://127.0.0.1:8118")
	// transport.Proxy = http.ProxyURL(u)
	cfg.WithHTTPClient(&http.Client{Transport: transport})
	cfg.WithUseSSL(true)

	return cfg
}
