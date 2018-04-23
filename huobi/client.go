package huobi

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/gotoxu/log/core"
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

// WSSClient 是huobi sdk的调用客户端
type WSSClient struct {
	config config.Config
	conns  map[string]*websocket.Conn
	logger core.Logger

	closed  bool
	closeMu sync.Mutex

	shouldQuit chan struct{}
	done       chan struct{}
	retry      chan string
}

// NewWSSClient 创建一个新的websocket客户端
func NewWSSClient(config *config.Config) *WSSClient {
	cfg := defaultConfig()
	if config != nil {
		cfg.MergeIn(config)
	}

	return &WSSClient{
		config:     *cfg,
		conns:      make(map[string]*websocket.Conn),
		shouldQuit: make(chan struct{}),
		done:       make(chan struct{}),
		retry:      make(chan string),
	}
}

// SetLogger 设置日志器
func (c *WSSClient) SetLogger(logger core.Logger) {
	c.logger = logger
}

func (c *WSSClient) log(level core.Level, v ...interface{}) {
	if c.logger != nil {
		c.logger.Log(level, v...)
	}
}

func (c *WSSClient) logf(level core.Level, format string, v ...interface{}) {
	if c.logger != nil {
		c.logger.Logf(level, format, v...)
	}
}

func (c *WSSClient) logln(level core.Level, v ...interface{}) {
	if c.logger != nil {
		c.logger.Logln(level, v...)
	}
}

// Close 发起关闭操作
func (c *WSSClient) Close() {
	if c.conns == nil || len(c.conns) == 0 {
		return
	}

	close(c.shouldQuit)

	select {
	case <-c.done:
	case <-time.After(time.Second):
	}
}

func (c *WSSClient) start(req interface{}, cid string, msgCh chan<- []byte) {
	go c.query(cid, msgCh)

	for {
		select {
		case cid := <-c.retry:
			delete(c.conns, cid)
			c.reconnect(req, msgCh)
		case <-c.shouldQuit:
			c.shutdown()
			return
		}
	}
}

func (c *WSSClient) reconnect(req interface{}, msgCh chan<- []byte) {
	cid, conn, err := c.connect()
	if err != nil {
		return
	}

	err = conn.WriteJSON(req)
	if err != nil {
		c.closeMu.Lock()
		defer c.closeMu.Unlock()

		if !c.closed {
			c.retry <- cid
		}
		return
	}

	go c.query(cid, msgCh)
}

func (c *WSSClient) shutdown() {
	c.closeMu.Lock()
	c.closed = true
	c.closeMu.Unlock()

	for _, conn := range c.conns {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		conn.Close()
	}

	close(c.done)
}

func (c *WSSClient) query(cid string, msgCh chan<- []byte) {
	conn, ok := c.conns[cid]
	if !ok {
		return
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			c.closeMu.Lock()
			defer c.closeMu.Unlock()

			if !c.closed {
				c.retry <- cid
			}

			return
		}

		buf := bytes.NewBuffer(msg)
		gz, err := gzip.NewReader(buf)
		if err != nil {
			continue
		}
		message, _ := ioutil.ReadAll(gz)

		if strings.Contains(string(message), "ping") {
			c.pong(cid, conn, message)
			continue
		}

		msgCh <- message
	}
}

func (c *WSSClient) pong(cid string, conn *websocket.Conn, msg []byte) {
	var ping struct {
		Ping int64 `json:"ping"`
	}

	err := json.Unmarshal(msg, &ping)
	if err != nil {
		c.closeMu.Lock()
		defer c.closeMu.Unlock()

		if !c.closed {
			c.retry <- cid
		}

		return
	}

	pong := struct {
		Pong int64 `json:"pong"`
	}{ping.Ping}
	conn.WriteJSON(pong)
}

func (c *WSSClient) connect() (string, *websocket.Conn, error) {
	u := url.URL{Scheme: "wss", Host: *c.config.WSSHost, Path: "/ws"}
	conn, _, err := c.config.WSSDialer.Dial(u.String(), nil)
	if err == nil {
		u := uuid.New().String()
		c.conns[u] = conn
		return u, conn, nil
	}

	if err == websocket.ErrBadHandshake {
		return "", nil, err
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			conn, _, err := c.config.WSSDialer.Dial(u.String(), nil)
			if err == nil {
				u := uuid.New().String()
				c.conns[u] = conn
				return u, conn, nil
			}
		case <-c.shouldQuit:
			return "", nil, errors.New("Connection is closing")
		}
	}
}

func (c *WSSClient) generateClientID() string {
	now := time.Now().UnixNano()
	return strconv.FormatInt(now, 10)
}

// Client 提供火币 API的调用客户端
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
	if e != nil {
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

	fmt.Printf("huobi http message:%s\n", string(body))
	extra.RegisterFuzzyDecoders()

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
