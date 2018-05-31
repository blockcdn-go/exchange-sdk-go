package gate

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/blockcdn-go/exchange-sdk-go/global"
	"github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

// GetAllSymbol 交易市场详细行情接口
func (c *Client) GetAllSymbol() ([]global.TradeSymbol, error) {
	var result struct {
		Result string               `json:"result"`
		Data   []MarketListResponse `json:"data"`
	}
	e := c.httpReq("GET", "/api2/1/marketlist", nil, &result)
	if e != nil {
		return nil, e
	}
	r := []global.TradeSymbol{}
	for _, s := range result.Data {
		r = append(r, global.TradeSymbol{
			Base:  s.CurrA,
			Quote: s.CurrB,
		})
	}
	return r, nil
}

// TickerInfo 获取行情ticker
func (c *Client) TickerInfo(base, quote string) (TickerResponse, error) {
	path := fmt.Sprintf("/api2/1/ticker/%s_%s", base, quote)
	var result TickerResponse
	e := c.httpReq("GET", path, nil, &result)
	result.Base = base
	result.Quote = quote
	return result, e
}

// DepthInfo 获取深度行情
func (c *Client) DepthInfo(base, quote string) (Depth5, error) {
	path := fmt.Sprintf("/api2/1/orderBook/%s_%s", base, quote)
	t := struct {
		Asks [][]float64 `json:"asks"` //卖方深度
		Bids [][]float64 `json:"bids"` //买方深度
	}{}

	e := c.httpReq("GET", path, nil, &t)
	if e != nil {
		return Depth5{}, e
	}
	if len(t.Asks) < 5 || len(t.Bids) < 5 {
		return Depth5{}, fmt.Errorf("depth len < 5")
	}
	var r Depth5
	r.Base = base
	r.Quote = quote
	r.Asks = make([]PSpair, 0, 5)
	r.Bids = make([]PSpair, 0, 5)
	if len(t.Asks) >= 2 && t.Asks[0][0] > t.Asks[1][0] {
		// 卖 倒序
		for end := len(t.Asks); end > len(t.Asks)-5; end-- {
			r.Asks = append(r.Asks, PSpair{t.Asks[end-1][0], t.Asks[end-1][1]})
		}
	} else {
		for i := 0; i < 5; i++ {
			r.Asks = append(r.Asks, PSpair{t.Asks[i][0], t.Asks[i][1]})
		}
	}

	// 买
	for i := 0; i < 5; i++ {
		r.Bids = append(r.Bids, PSpair{t.Bids[i][0], t.Bids[i][1]})
	}
	return r, nil
}

// GetKline 获取k线数据
func (c *Client) GetKline(req global.KlineReq) ([]global.Kline, error) {
	groupSec := 60
	rangeHour := 1
	if req.Period == "5m" {
		groupSec = 300
		rangeHour = 12
	} else if req.Period == "15m" {
		groupSec = 900
		rangeHour = 24
	} else if req.Period == "30m" {
		groupSec = 1800
		rangeHour = 48
	} else if req.Period == "1h" {
		groupSec = 3600
		rangeHour = 96
	} else if req.Period == "8h" {
		groupSec = 28800
		rangeHour = 768
	} else if req.Period == "1d" {
		groupSec = 86400
		rangeHour = 2304
	}
	sym := strings.ToLower(req.Base + "_" + req.Quote)
	path := fmt.Sprintf("/api2/1/candlestick2/%s?group_sec=%d&range_hour=%d", sym, groupSec, rangeHour)
	rsp := struct {
		Result  string      `json:"result"`
		Message string      `json:"message"`
		Code    int64       `json:"code"`
		Data    [][]float64 `json:"data"`
	}{}

	e := c.httpReq("GET", path, nil, &rsp)
	if e != nil {
		return nil, e
	}
	if rsp.Result != "true" {
		return nil, fmt.Errorf(rsp.Message)
	}
	k := []global.Kline{}
	for i := 0; i < len(rsp.Data); i++ {
		if len(rsp.Data[i]) < 6 {
			fmt.Println("gate len(rsp.Data[i]) < 6")
			continue
		}
		k = append(k, global.Kline{
			Timestamp: int64(rsp.Data[i][0]),
			Volume:    rsp.Data[i][1],
			Close:     rsp.Data[i][2],
			High:      rsp.Data[i][3],
			Low:       rsp.Data[i][4],
			Open:      rsp.Data[i][5],
		})
	}
	return k, nil
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

// GetFund 获取帐号资金余额
func (c *Client) GetFund(global.FundReq) ([]global.Fund, error) {
	path := "/api2/1/private/balances"
	b := struct {
		Result    string            `json:"result"`
		Available map[string]string `json:"available"`
		Locked    map[string]string `json:"locked"`
	}{}
	e := c.httpReq("POST", path, nil, &b)
	if e != nil {
		return nil, e
	}
	if b.Result != "true" {
		return nil, fmt.Errorf("get balances result false")
	}
	var r Balance
	r.Available = make(map[string]float64)
	r.Locked = make(map[string]float64)

	if ce := convkv(r.Available, b.Available); ce != nil {
		return nil, ce
	}
	if ce := convkv(r.Locked, b.Locked); ce != nil {
		return nil, ce
	}

	ir := []global.Fund{}
	for k, av := range r.Available {
		ir = append(ir, global.Fund{
			Base:      k,
			Available: av,
		})
	}

	for k, fz := range r.Locked {
		find := false
		for i := 0; i < len(ir); i++ {
			if ir[i].Base == k {
				ir[i].Frozen += fz
				find = true
				break
			}
		}
		if !find {
			ir = append(ir, global.Fund{
				Base:   k,
				Frozen: fz,
			})
		}
	}
	return ir, nil
}

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

// InsertOrder 下单交易
// @parm symbol 交易币种对(如ltc_btc,ltc_btc)
// @parm direction 0 - buy, 1 - sell
// @parm price 	买卖价格 ps: minimum 10 usdt.
// @parm num	买卖币数量
func (c *Client) InsertOrder(req global.InsertReq) (global.InsertRsp, error) {
	path := "/api2/1/private/"
	if req.Direction == 0 {
		path += "buy"
	} else {
		path += "sell"
	}
	symbol := strings.ToLower(req.Base + "_" + req.Quote)
	arg := struct {
		CurrencyPair string  `url:"currencyPair"`
		Rate         float64 `url:"rate"`
		Amount       float64 `url:"amount"`
	}{symbol, req.Price, req.Num}
	r := InsertOrderRsp{Direction: req.Direction}
	e := c.httpReq("POST", path, arg, &r)
	if e != nil {
		return global.InsertRsp{}, e
	}
	if r.Result != "true" {
		return global.InsertRsp{}, fmt.Errorf("gateio error: %s", r.Msg)
	}
	return global.InsertRsp{OrderNo: r.OrderNo}, e
}

// CancelOrder 取消订单
// 通过测试，第一个参数对结果没有影响，只要orderno正确就能取消订单，
// 但是如果第一个参数填入错误的代码将返回错误，但是订单依然被取消了
func (c *Client) CancelOrder(req global.CancelReq) error {
	symbol := strings.ToLower(req.Base + "_" + req.Quote)
	arg := struct {
		OrderNumber  string `url:"orderNumber"`
		CurrencyPair string `url:"currencyPair"`
	}{req.OrderNo, symbol}

	r := struct {
		Result  interface{} `json:"result"` // 未按文档说明的类型返回
		BResult bool        `json:"-"`
		Code    int         `json:"code"`
		Message string      `json:"message"`
	}{}
	e := c.httpReq("POST", "/api2/1/private/cancelOrder", arg, &r)
	if e != nil {
		return e
	}

	switch r.Result.(type) {
	case bool:
		r.BResult = r.Result.(bool)
	case string:
		v := r.Result.(string)
		r.BResult, _ = strconv.ParseBool(v)
	default:
		r.BResult = false
	}

	if !r.BResult && r.Code != 0 {
		return fmt.Errorf(r.Message)
	}
	return nil
}

// OrderStatus 获取订单状态
func (c *Client) OrderStatus(req global.StatusReq) (global.StatusRsp, error) {
	symbol := strings.ToLower(req.Base + "_" + req.Quote)
	arg := struct {
		OrderNumber  string `url:"orderNumber"`
		CurrencyPair string `url:"currencyPair"`
	}{req.OrderNo, symbol}
	r := struct {
		Result  string    `json:"result"`
		Message string    `json:"message"`
		Order   OrderInfo `json:"order"`
	}{}
	e := c.httpReq("POST", "/api2/1/private/getOrder", arg, &r)
	if e != nil {
		return global.StatusRsp{}, e
	}
	if r.Result != "true" {
		return global.StatusRsp{}, fmt.Errorf(r.Message)
	}

	or := &r.Order
	m := global.StatusRsp{}
	n, e := strconv.ParseFloat(or.TradeNum, 64)
	if e != nil {
		return m, e
	}
	m.TradePrice = or.TradePrice
	m.TradeNum = n
	if n != 0. {
		m.Status = global.HALFTRADE
		m.StatusMsg = "部分成交"
	}
	if or.Status == "done" {
		m.Status = global.COMPLETETRADE
		m.StatusMsg = "完全成交"
	}
	if or.Status == "cancelled" {
		m.Status = global.CANCELED
		m.StatusMsg = "已撤单"
	}
	fmt.Printf("gateio order status %+v\n", or)
	return m, nil
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

	//fmt.Printf("http message: %s\n", string(body))
	extra.RegisterFuzzyDecoders()

	err = jsoniter.Unmarshal(body, out)
	if err != nil {
		return err
	}
	return nil
}
