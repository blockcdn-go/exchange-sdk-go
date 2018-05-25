package gate

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/gotoxu/log/core"
	"github.com/gotoxu/query"
)

// Client 提供gate API的调用客户端
type Client struct {
	config config.Config
	logger core.Logger
}

// NewClient 创建一个新的client
func NewClient(config *config.Config) *Client {
	cfg := defaultConfig()
	if config != nil {
		cfg.MergeIn(config)
	}

	return &Client{config: *cfg}
}

// SetLogger 设置日志器
func (c *Client) SetLogger(logger core.Logger) {
	c.logger = logger
}

func (c *Client) log(level core.Level, v ...interface{}) {
	if c.logger != nil {
		c.logger.Log(level, v...)
	}
}

func (c *Client) logf(level core.Level, format string, v ...interface{}) {
	if c.logger != nil {
		c.logger.Logf(level, format, v...)
	}
}

func (c *Client) logln(level core.Level, v ...interface{}) {
	if c.logger != nil {
		c.logger.Logln(level, v...)
	}
}

func (c *Client) newRequest(method, endpoint, path string) *request {
	r := &request{
		config: &c.config,
		method: method,
		params: make(map[string][]string),
		header: make(http.Header),
		ctx:    c.config.Context,
	}

	u := &url.URL{
		Host: endpoint,
		Path: path,
	}

	if *c.config.UseSSL {
		u.Scheme = "https"
	} else {
		u.Scheme = "http"
	}

	r.url = u
	return r
}

func (c *Client) doRequest(r *request) (*http.Response, error) {
	req, err := r.toHTTP()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(netw, addr, time.Second*5) //设置建立连接超时
				if err != nil {
					return nil, err
				}
				conn.SetDeadline(time.Now().Add(time.Second * 5)) //设置发送接受数据超时
				return conn, nil
			},
			ResponseHeaderTimeout: time.Second * 5,
		},
	}
	return client.Do(req)
	//return c.config.HTTPClient.Do(req)
}

func (c *Client) encodeFormBody(obj interface{}) (io.Reader, string, error) {
	encoder := query.NewEncoder()

	var err error
	var form url.Values
	form, err = encoder.Encode(obj)

	if err != nil {
		return nil, "", err
	}

	return strings.NewReader(form.Encode()), form.Encode(), nil
}

func (c *Client) sign(params string) string {
	key := []byte(*c.config.Secret)
	mac := hmac.New(sha512.New, key)
	mac.Write([]byte(params))

	hash := mac.Sum(nil)
	return hex.EncodeToString(hash)
}

type request struct {
	config *config.Config
	method string
	url    *url.URL
	params url.Values
	body   io.Reader
	header http.Header
	sign   string
	ctx    context.Context
}

func (r *request) toHTTP() (*http.Request, error) {
	r.url.RawQuery = r.params.Encode()

	req, err := http.NewRequest(r.method, r.url.RequestURI(), r.body)
	if err != nil {
		return nil, err
	}

	req.URL = r.url
	req.Header = r.header

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if r.sign != "" {
		req.Header.Set("key", *r.config.APIKey)
		req.Header.Set("sign", r.sign)
	}

	if r.ctx != nil {
		return req.WithContext(r.ctx), nil
	}

	return req, nil
}
