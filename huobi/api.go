package huobi

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/blockcdn-go/exchange-sdk-go/global"
	"gitlab.mybcdn.com/golang/blockcoin/apidb"
)

// GetAllSymbol 获取所有的可交易对
func (c *Client) GetAllSymbol() ([]global.TradeSymbol, error) {
	r := struct {
		Status string      `json:"status"`
		Data   []TradePair `json:"data"`
		Errmsg string      `json:"err-msg"`
	}{}
	e := c.doHTTP("GET", "/v1/common/symbols", nil, &r)
	if e != nil {
		return nil, e
	}
	if r.Status != "ok" {
		return nil, fmt.Errorf(r.Errmsg)
	}
	ir := []global.TradeSymbol{}
	for _, p := range r.Data {
		ir = append(ir, global.TradeSymbol{
			Base:  p.Base,
			Quote: p.Quote,
		})
	}
	return ir, nil
}

// GetAllAccountID 获取用户的所有accountid
// GET /v1/account/accounts 查询当前用户的所有账户(即account-id)，Pro站和HADAX account-id通用
func (c *Client) GetAllAccountID() ([]Account, error) {

	r := struct {
		Status string    `json:"status"`
		Data   []Account `json:"data"`
		Errmsg string    `json:"err-msg"`
	}{}
	e := c.doHTTP("GET", "/v1/account/accounts", nil, &r)
	if e != nil {
		return nil, e
	}
	if r.Status != "ok" {
		return nil, fmt.Errorf(r.Errmsg)
	}
	return r.Data, nil
}

// GetKline 获取k线数据
func (c *Client) GetKline(req global.KlineReq) ([]global.Kline, error) {

	period := req.Period
	if strings.Contains(period, "m") {
		period = period + "in"
	} else if period == "1h" {
		period = "60m"
	} else if strings.Contains(period, "d") {
		period = period + "ay"
	} else if strings.Contains(period, "w") {
		period = period + "eek"
	}
	arg := make(map[string]string)
	arg["symbol"] = strings.ToLower(req.Base + req.Quote)
	if req.Count != 0 {
		arg["size"] = strconv.FormatInt(req.Count, 10)
	}
	arg["period"] = period

	r := struct {
		Status string  `json:"status"`
		Ch     string  `json:"ch"`
		Data   []Kline `json:"data"`
		Errmsg string  `json:"err-msg"`
	}{}
	e := c.doHTTP("GET", "/market/history/kline", arg, &r)
	if e != nil {
		return nil, e
	}
	if r.Status != "ok" {
		return nil, fmt.Errorf(r.Errmsg)
	}
	ik := []global.Kline{}
	for _, k := range r.Data {
		ik = append(ik, global.Kline{
			Base:      k.Base,
			Quote:     k.Quote,
			Timestamp: int64(k.Timestamp),
			High:      k.High,
			Open:      k.Open,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
		})
	}
	return ik, nil
}

// GetFund 查询指定账户的余额
// @parm accountID GetAllAccountID函数返回的id
func (c *Client) GetFund(req global.FundReq) ([]global.Fund, error) {
	ids, e := c.GetAllAccountID()
	if e != nil {
		return nil, e
	}
	if len(ids) == 0 {
		return nil, errors.New("huobipro no accountid")
	}
	r := struct {
		Status string `json:"status"`
		Data   struct {
			List []Balance `json:"list"`
		} `json:"data"`

		Errmsg string `json:"err-msg"`
	}{}

	path := fmt.Sprintf("/v1/account/accounts/%d/balance", ids[0].AccountID)
	e = c.doHTTP("GET", path, nil, &r)
	if e != nil {
		return nil, e
	}
	if r.Status != "ok" {
		return nil, fmt.Errorf(r.Errmsg)
	}

	ir := []global.Fund{}
	for _, bb := range r.Data.List {
		t := global.Fund{
			Base: bb.CurrencyNo,
		}
		if bb.BType == "trade" {
			t.Available, _ = strconv.ParseFloat(bb.Amount, 64)
		} else if bb.BType == "frozen" {
			t.Frozen, _ = strconv.ParseFloat(bb.Amount, 64)
		} else {
			log.Println("火币账户资金类型错误")
			return nil, errors.New("火币账户资金类型错误")
		}
		find := false
		for i := 0; i < len(ir); i++ {
			if ir[i].Base == t.Base {
				find = true
				ir[i].Available += t.Available
				ir[i].Frozen += t.Frozen
				break
			}
		}
		if !find {
			ir = append(ir, t)
		}
	}
	return ir, nil
}

// InsertOrder 下单
// @return string: orderNo
func (c *Client) InsertOrder(req global.InsertReq) (global.InsertRsp, error) {
	ids, e := c.GetAllAccountID()
	if e != nil {
		return global.InsertRsp{}, e
	}
	if len(ids) == 0 {
		return global.InsertRsp{}, errors.New("huobipro no accountid")
	}

	ireq := InsertOrderReq{
		Source:    "api",
		AccountID: strconv.FormatInt(ids[0].AccountID, 10),
		Price:     strconv.FormatFloat(req.Price, 'f', -1, 64),
		Amount:    strconv.FormatFloat(req.Num, 'f', -1, 64),
		Symbol:    strings.ToLower(req.Base + req.Quote),
	}
	sd := "buy"
	st := "limit"
	if apidb.OrderDirection(req.Direction) == apidb.SELL {
		sd = "sell"
	}
	if apidb.OrderType(req.Type) == apidb.MARKET {
		st = "market"
	}
	ireq.OrderType = sd + "-" + st

	mapParams := if2map(req)
	r := struct {
		Status string `json:"status"`
		Errmsg string `json:"err-msg"`
		Data   string `json:"data"`
	}{}
	e = c.doHTTP("POST", "/v1/hadax/order/orders/place", mapParams, &r)
	if e != nil {
		return global.InsertRsp{}, e
	}
	if r.Status != "ok" {
		return global.InsertRsp{}, fmt.Errorf(r.Errmsg)
	}
	return global.InsertRsp{OrderNo: r.Data}, nil
}

// CancelOrder 撤销一个订单请求
// 注意，返回OK表示撤单请求成功。订单是否撤销成功请调用订单查询接口查询该订单状态
func (c *Client) CancelOrder(req global.CancelReq) error {
	path := fmt.Sprintf("/v1/order/orders/%s/submitcancel", req.OrderNo)
	r := struct {
		Status string `json:"status"`
		Errmsg string `json:"err-msg"`
		Data   string `json:"data"`
	}{}
	e := c.doHTTP("POST", path, nil, &r)
	if e != nil {
		return e
	}
	if r.Status != "ok" {
		return fmt.Errorf(r.Errmsg)
	}
	return nil
}

// OrderStatus 查询某个订单详情
func (c *Client) OrderStatus(req global.StatusReq) (global.StatusRsp, error) {
	path := fmt.Sprintf("/v1/order/orders/%s", req.OrderNo)
	r := struct {
		Status string      `json:"status"`
		Errmsg string      `json:"err-msg"`
		Data   OrderDetail `json:"data"`
	}{}
	e := c.doHTTP("GET", path, nil, &r)
	if e != nil {
		return global.StatusRsp{}, e
	}
	if r.Status != "ok" {
		return global.StatusRsp{}, fmt.Errorf(r.Errmsg)
	}
	or := &r.Data
	m := global.StatusRsp{}
	m.TradePrice, _ = strconv.ParseFloat(or.TradePrice, 64)
	m.TradeNum, _ = strconv.ParseFloat(or.TradeNum, 64)
	if m.TradeNum != 0. || or.OrderStatus == "partial-canceled" {
		m.Status = global.HALFTRADE
		m.StatusMsg = "部分成交"
	}
	if or.OrderStatus == "filled" {
		m.Status = global.COMPLETETRADE
		m.StatusMsg = "完全成交"
	}
	if or.OrderStatus == "canceled" {
		m.Status = global.CANCELED
		m.StatusMsg = "已撤单"
	}
	fmt.Printf("huobipro order status %+v\n", or)
	return m, nil
}

// GetMatchDetail 查询某个订单的成交明细
func (c *Client) GetMatchDetail(orderno string) ([]MatchDetail, error) {
	path := fmt.Sprintf("/v1/order/orders/%s/matchresults", orderno)
	r := struct {
		Status string        `json:"status"`
		Errmsg string        `json:"err-msg"`
		Data   []MatchDetail `json:"data"`
	}{}
	e := c.doHTTP("GET", path, nil, &r)
	if e != nil {
		return nil, e
	}
	if r.Status != "ok" {
		return nil, fmt.Errorf(r.Errmsg)
	}
	return r.Data, nil
}

// GetOrders 查询当前委托、历史委托
// @parm symbol 交易对		btcusdt, bchbtc, rcneth ...
// @parm status 查询的订单状态组合，使用','分割
//				pre-submitted 准备提交,
//				submitted 已提交,
//				partial-filled 部分成交,
//				partial-canceled 部分成交撤销,
//				filled 完全成交,
//				canceled 已撤销
func (c *Client) GetOrders(symbol, status string) ([]OrderDetail, error) {
	arg := make(map[string]string)
	arg["symbol"] = symbol
	arg["states"] = status
	r := struct {
		Status string        `json:"status"`
		Errmsg string        `json:"err-msg"`
		Data   []OrderDetail `json:"data"`
	}{}
	e := c.doHTTP("GET", "/v1/order/orders", arg, &r)
	if e != nil {
		return nil, e
	}
	if r.Status != "ok" {
		return nil, fmt.Errorf(r.Errmsg)
	}
	return r.Data, nil
}

// TODO: 全部撤单
// TODO: 查询当前成交、历史成交
