package huobi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/global"

	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
)

// WSSClient 是huobi sdk的调用客户端
type Client struct {
	config    config.Config
	replay    bool
	once      sync.Once
	sock      *websocket.Conn
	mutex     sync.Mutex
	tick      map[global.TradeSymbol]chan global.Ticker
	depth     map[global.TradeSymbol]chan global.Depth
	latetrade map[global.TradeSymbol]chan global.LateTrade
}

// NewClient 创建一个新的websocket客户端
func NewClient(config *config.Config) *Client {
	cfg := defaultConfig()
	if config != nil {
		cfg.MergeIn(config)
	}

	return &Client{
		config:    *cfg,
		tick:      make(map[global.TradeSymbol]chan global.Ticker),
		depth:     make(map[global.TradeSymbol]chan global.Depth),
		latetrade: make(map[global.TradeSymbol]chan global.LateTrade),
	}
}

func (c *Client) generateClientID() string {
	now := time.Now().UnixNano()
	return strconv.FormatInt(now, 10)
}

func (c *Client) doHTTP(method, path string, mapParams map[string]string, out interface{}) error {

	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05")

	mapParams2Sign := make(map[string]string)
	mapParams2Sign["AccessKeyId"] = *c.config.APIKey
	mapParams2Sign["SignatureMethod"] = "HmacSHA256"
	mapParams2Sign["SignatureVersion"] = "2"
	mapParams2Sign["Timestamp"] = timestamp
	for k, v := range mapParams {
		mapParams2Sign[k] = v
	}
	hostName := *c.config.RESTHost

	mapParams2Sign["Signature"] = createSign(mapParams2Sign, method, hostName,
		path, *c.config.Secret)

	url := "http://"
	if *c.config.UseSSL {
		url = "https://"
	}
	url += *c.config.RESTHost + path
	url += "?" + map2UrlQuery(mapParams2Sign)

	arg := ""
	if method == "POST" {
		bytesParams, _ := json.Marshal(mapParams)
		arg = string(bytesParams)
	}
	req, e := http.NewRequest(method, url, strings.NewReader(arg))
	if e != nil {
		return e
	}
	if method == "GET" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept-Language", "zh-cn")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.71 Safari/537.36")

	resp, re := c.config.HTTPClient.Do(req)
	if re != nil {
		return re
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，响应码：%d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if !strings.Contains(path, "/v1/common/symbols") &&
		!strings.Contains(path, "/market/history/kline") &&
		!strings.Contains(path, "/v1/account/accounts") &&
		!strings.Contains(path, "/v1/order/orders") {
		fmt.Printf("huobi http message:%s\n", string(body))
	}

	//extra.RegisterFuzzyDecoders()

	err = jsoniter.Unmarshal(body, out)
	if err != nil {
		return err
	}
	return nil
}

// 构造签名
func createSign(mapParams map[string]string, strMethod, strHostURL,
	strRequestPath, strSecretKey string) string {
	// 参数处理, 按API要求, 参数名应按ASCII码进行排序(使用UTF-8编码, 其进行URI编码, 16进制字符必须大写)
	strParams := valURIQuery(mapSort(mapParams))

	strPayload := strMethod + "\n" +
		strHostURL + "\n" +
		strRequestPath + "\n" +
		strParams

	key := []byte(strSecretKey)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(strPayload))

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func map2UrlQuery(mapParams map[string]string) string {
	var strParams string
	for key, value := range mapParams {
		strParams += (key + "=" + url.QueryEscape(value) + "&")
	}
	if 0 < len(strParams) {
		strParams = string([]rune(strParams)[:len(strParams)-1])
	}
	return strParams
}
