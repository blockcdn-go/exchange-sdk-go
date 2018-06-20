package aicoin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/blockcdn-go/exchange-sdk-go/utils"
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

// Client 提供 API的调用客户端
type Client struct {
	Config config.Config
}

// Constructor 创建一个新的client
func (c *Client) Constructor(config *config.Config) {
	cfg := defaultConfig()
	if config != nil {
		cfg.MergeIn(config)
	}
	extra.RegisterFuzzyDecoders()
	c.Config = *cfg
}

func (c *Client) aicoinHTTPReq(method, path string, in map[string]interface{}, out interface{}) error {
	if in == nil {
		in = make(map[string]interface{})
	}

	rbody, _ := json.Marshal(in)
	if method == "GET" {
		rbody = []byte{}
	}
	path += "?" + utils.MapEncode(in)

	req, err := http.NewRequest(method, path, bytes.NewReader(rbody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36")

	req.Header.Set("Referer", "https://www.aicoin.net.cn/")
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
	fmt.Printf("aicoin message: %s\n", string(body))
	//extra.RegisterFuzzyDecoders()

	err = jsoniter.Unmarshal(body, out)
	if err != nil {
		return err
	}
	return nil
}
