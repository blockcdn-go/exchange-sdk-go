package gate

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"io/ioutil"

	"github.com/blockcdn-go/exchange-sdk-go/config"
)

// Client 提供gate API的调用客户端
type Client struct {
	config config.Config
}

// NewClient 创建一个新的client
func NewClient(config *config.Config) *Client {
	cfg := defaultConfig()
	if config != nil {
		cfg.MergeIn(config)
	}

	return &Client{config: *cfg}
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

	return c.config.HTTPClient.Do(req)
}

type request struct {
	config *config.Config
	method string
	url    *url.URL
	params url.Values
	body   io.Reader
	header http.Header
	ctx    context.Context
}

func (r *request) toHTTP() (*http.Request, error) {
	req, err := http.NewRequest(r.method, r.url.RequestURI(), r.body)
	if err != nil {
		return nil, err
	}

	req.URL = r.url
	req.Header = r.header

	var str string
	if r.body != nil {
		p, e := ioutil.ReadAll(r.body)
		if e == nil {
			str = string(p)
		}
	}
	//sign := r.sign(r.params.Encode())
	sign := r.sign(str)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("key", *r.config.APIKey)
	req.Header.Set("sign", sign)

	if r.ctx != nil {
		return req.WithContext(r.ctx), nil
	}

	return req, nil
}

func (r *request) sign(params string) string {
	key := []byte(*r.config.Secret)
	mac := hmac.New(sha512.New, key)
	mac.Write([]byte(params))

	hash := mac.Sum(nil)
	return hex.EncodeToString(hash)
}
