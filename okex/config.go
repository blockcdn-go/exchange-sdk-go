package okex

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Config 是okex sdk的相关配置
type Config struct {
	RESTHost   *string
	WSSHost    *string
	APIKey     *string
	Secret     *string
	UseSSL     *bool
	HTTPClient *http.Client
	WSSDialer  *websocket.Dialer
	Context    context.Context
}

// WithAPIKey 设置sdk访问的API key
func (c *Config) WithAPIKey(key string) *Config {
	c.APIKey = &key
	return c
}

// WithSecret 设置sdk访问的Secret
func (c *Config) WithSecret(secret string) *Config {
	c.Secret = &secret
	return c
}

// WithUseSSL 设置sdk访问rest接口时是否使用https
func (c *Config) WithUseSSL(use bool) *Config {
	c.UseSSL = &use
	return c
}

// WithHTTPClient 设置请求rest接口时的http client
func (c *Config) WithHTTPClient(client *http.Client) *Config {
	c.HTTPClient = client
	return c
}

// WithWSSDialer 设置自定义的Websocket dialer
func (c *Config) WithWSSDialer(dialer *websocket.Dialer) *Config {
	c.WSSDialer = dialer
	return c
}

// WithContext 设置自定义context.Context
func (c *Config) WithContext(ctx context.Context) *Config {
	c.Context = ctx
	return c
}

// WithRESTHost 设置rest接口的地址
func (c *Config) WithRESTHost(host string) *Config {
	c.RESTHost = &host
	return c
}

// WithWSSHost 设置wss接口的地址
func (c *Config) WithWSSHost(host string) *Config {
	c.WSSHost = &host
	return c
}

// MergeIn 用于合并多个配置
func (c *Config) MergeIn(cfgs ...*Config) {
	for _, other := range cfgs {
		mergeInConfig(c, other)
	}
}

func mergeInConfig(dst *Config, other *Config) {
	if other == nil {
		return
	}

	if other.APIKey != nil {
		dst.APIKey = other.APIKey
	}
	if other.UseSSL != nil {
		dst.UseSSL = other.UseSSL
	}
	if other.Secret != nil {
		dst.Secret = other.Secret
	}
	if other.RESTHost != nil {
		dst.RESTHost = other.RESTHost
	}
	if other.WSSHost != nil {
		dst.WSSHost = other.WSSHost
	}
	if other.HTTPClient != nil {
		dst.HTTPClient = other.HTTPClient
	}
	if other.WSSDialer != nil {
		dst.WSSDialer = other.WSSDialer
	}
	if other.Context != nil {
		dst.Context = other.Context
	}
}

// DefaultConfig 返回默认sdk配置
func DefaultConfig() *Config {
	cfg := &Config{}
	// todo: 完善默认配置
	cfg.WithRESTHost("")
	cfg.WithWSSHost("okexcomreal.bafang.com:10441")
	cfg.WithHTTPClient(defaultHTTPClient())
	cfg.WithWSSDialer(defaultWSSDialer())
	cfg.WithUseSSL(true)

	return cfg
}

func defaultHTTPClient() *http.Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:        100,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return &http.Client{Transport: transport}
}

func defaultWSSDialer() *websocket.Dialer {
	return websocket.DefaultDialer
}
