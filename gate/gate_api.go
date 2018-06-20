package gate

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/json-iterator/go"
)

// TickerInfo 获取行情ticker
func (c *Client) TickerInfo(base, quote string) (TickerResponse, error) {
	path := fmt.Sprintf("/api2/1/ticker/%s_%s", base, quote)
	var result TickerResponse
	e := c.httpReq("GET", path, nil, &result)
	result.Base = base
	result.Quote = quote
	return result, e
}

// LateTradeInfo 获取最近80条成交
func (c *Client) LateTradeInfo(base, quote string) ([]LateTrade, error) {
	path := fmt.Sprintf("/api2/1/tradeHistory/%s_%s", base, quote)
	rsp := struct {
		Result  string      `json:"result"`
		Message string      `json:"message"`
		Code    int64       `json:"code"`
		Data    []LateTrade `json:"data"`
	}{}

	e := c.httpReq("GET", path, nil, &rsp)
	if e != nil {
		return nil, e
	}
	if rsp.Result != "true" {
		return nil, fmt.Errorf(rsp.Message)
	}
	for i := 0; i < len(rsp.Data); i++ {
		rsp.Data[i].Base = base
		rsp.Data[i].Quote = quote
	}
	return rsp.Data, nil
}

//////////////////////////////////////////////////////////////////////////
/// 交易类接口

// DepositAddr 获取充值地址
func (c *Client) DepositAddr(currency string) (string, error) {
	path := "/api2/1/private/depositAddress"
	rsp := struct {
		Result  string `json:"result"`
		Addr    string `json:"addr"`
		Message string `json:"message"`
		Code    int64  `json:"code"`
	}{}

	arg := struct {
		Currency string `url:"currency"`
	}{currency}

	e := c.httpReq("POST", path, arg, &rsp)
	if e != nil {
		return "", e
	}
	if rsp.Result != "true" && rsp.Code != 0 {
		return "", fmt.Errorf(rsp.Message)
	}
	return rsp.Addr, nil
}

// DepositsWithdrawals 获取充值提现历史
// return1 充值， return2 提现
func (c *Client) DepositsWithdrawals() ([]DWInfo, []DWInfo, error) {
	rsp := struct {
		Result    string   `json:"result"`
		Message   string   `json:"message"`
		Deposits  []DWInfo `json:"deposits"`
		Withdraws []DWInfo `json:"withdraws"`
	}{}
	e := c.httpReq("POST", "/api2/1/private/depositsWithdrawals", nil, &rsp)
	if e != nil {
		return nil, nil, e
	}
	if rsp.Result != "true" {
		return nil, nil, fmt.Errorf(rsp.Message)
	}
	return rsp.Deposits, rsp.Withdraws, nil
}

// HangingOrderInfo 获取我的当前挂单列表
func (c *Client) HangingOrderInfo() ([]HangingOrder, error) {
	r := struct {
		Result  string         `json:"result"`
		Message string         `json:"message"`
		Code    int            `json:"code"`
		Orders  []HangingOrder `json:"orders"`
	}{}
	e := c.httpReq("POST", "/api2/1/private/openOrders", nil, &r)
	if e != nil {
		return nil, e
	}
	if r.Result != "true" && r.Code != 0 {
		return nil, fmt.Errorf(r.Message)
	}
	return r.Orders, nil
}

// MatchInfo 获取我的24小时内成交记录
func (c *Client) MatchInfo(symbol, orderNo string) ([]Match, error) {
	arg := struct {
		OrderNumber  string `url:"orderNumber"`
		CurrencyPair string `url:"currencyPair"`
	}{orderNo, symbol}

	r := struct {
		Result  string  `json:"result"`
		Message string  `json:"message"`
		Trades  []Match `json:"trades"`
	}{}
	e := c.httpReq("POST", "/api2/1/private/tradeHistory", arg, &r)
	if e != nil {
		return nil, e
	}
	if r.Result != "true" {
		return nil, fmt.Errorf(r.Message)
	}
	return r.Trades, nil
}

// Withdraws 提现
func (c *Client) Withdraws(currency, address string, num float64) error {
	arg := struct {
		Currency string  `url:"currency"`
		Amount   float64 `url:"amount"`
		Address  string  `url:"address"`
	}{currency, num, address}

	r := struct {
		Result  string `json:"result"`
		Message string `json:"message"`
	}{}
	e := c.httpReq("POST", "/api2/1/private/withdraw", arg, &r)
	if e != nil {
		return e
	}
	if r.Result != "true" {
		return fmt.Errorf(r.Message)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////

func convkv(dst map[string]float64, src map[string]string) error {
	for k, v := range src {
		f, e := strconv.ParseFloat(v, 64)
		if e != nil {
			return e
		}
		dst[k] = f
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////
func (c *Client) httpReq(method, path string, in interface{}, out interface{}) error {
	r := c.newRequest(method, *c.config.RESTHost, path)
	if in != nil {
		body, params, err := c.encodeFormBody(in)
		if err != nil {
			return err
		}
		r.body = body
		r.sign = c.sign(params)
	} else {
		r.sign = c.sign("")
	}

	resp, err := c.doRequest(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，响应码：%d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if strings.Contains(path, "/api2/1/private/") && !strings.Contains(path, "getOrder") {
		fmt.Printf("http message: %s\n", string(body))
	}

	// extra.RegisterFuzzyDecoders()

	err = jsoniter.Unmarshal(body, out)
	if err != nil {
		return err
	}
	return nil
}
