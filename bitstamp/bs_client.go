package bitstamp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/baseclass"
	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/blockcdn-go/exchange-sdk-go/utils"
	jsoniter "github.com/json-iterator/go"
)

// Client 提供 API的调用客户端
type Client struct {
	baseclass.Client
}

// NewClient 创建一个新的client
func NewClient(config *config.Config) *Client {
	cfg := defaultConfig()
	if config != nil {
		cfg.MergeIn(config)
	}
	c := &Client{}
	c.Constructor(config)
	return c
}

func (c *Client) httpReq(method, path string, in map[string]interface{}, out interface{}, bs bool) error {
	if in == nil {
		in = make(map[string]interface{})
	}
	sig := ""
	if bs {
		in["key"] = *c.Config.APIKey
		in["nonce"] = utils.ToString(time.Now().Unix())
		sig = sign(utils.MapEncode(in), *c.Config.Secret)
		in["signature"] = sig
	}
	rbody, _ := json.Marshal(in)
	if method == "GET" {
		rbody = []byte{}
	}
	path += "?" + utils.MapEncode(in)

	fmt.Println(path)
	req, err := http.NewRequest(method, path, bytes.NewReader(rbody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.71 Safari/537.36")

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
