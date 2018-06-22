package coinex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/blockcdn-go/exchange-sdk-go/global"

	"github.com/blockcdn-go/exchange-sdk-go/baseclass"
	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/blockcdn-go/exchange-sdk-go/utils"
	jsoniter "github.com/json-iterator/go"
)

// Client 提供 API的调用客户端
type Client struct {
	baseclass.Client
	sock     *websocket.Conn
	tickOnce sync.Once
	mtx      sync.Mutex
	ltid     int64 // 最后一次成交的id
	tick     map[global.TradeSymbol]chan global.Ticker
}

// NewClient 创建一个新的client
func NewClient(config *config.Config) *Client {
	cfg := defaultConfig()
	if config != nil {
		cfg.MergeIn(config)
	}
	c := &Client{
		tick: make(map[global.TradeSymbol]chan global.Ticker),
	}
	c.Exchange = "coinex"
	c.Constructor(config)
	return c
}

func (c *Client) httpReq(method, path string, in map[string]interface{}, out interface{}, bs bool) error {
	if in == nil {
		in = make(map[string]interface{})
	}
	sig := ""
	urlPs := ""
	if bs {
		in["access_id"] = *c.Config.APIKey
		in["tonce"] = utils.ToString(time.Now().UnixNano() / 1000000)
		urlPs = utils.MapEncode(in) + "&secret_key=" + *c.Config.Secret
		sig = sign(urlPs, *c.Config.Secret)
	} else {
		urlPs = utils.MapEncode(in)
	}
	rbody, _ := json.Marshal(in)
	if method == "GET" {
		rbody = []byte{}
	}
	path += "?" + urlPs

	fmt.Println(path)
	req, err := http.NewRequest(method, path, bytes.NewReader(rbody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.71 Safari/537.36")
	if bs {
		req.Header.Set("authorization", sig)
	}
	resp, err := c.Config.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，响应码：%d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("http message: %s\n", string(body))
	//extra.RegisterFuzzyDecoders()

	err = jsoniter.Unmarshal(body, out)
	if err != nil {
		return err
	}
	return nil
}
