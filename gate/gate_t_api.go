package gate

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

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

// InsertOrder 下单交易
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
	n, e := strconv.ParseFloat(or.InsertNum, 64)
	if e != nil {
		return m, e
	}
	m.TradePrice = or.TradePrice
	m.TradeNum = n
	if or.Status == "closed" {
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
