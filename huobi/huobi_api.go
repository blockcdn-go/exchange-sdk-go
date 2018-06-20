package huobi

import (
	"fmt"
)

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
// func (c *Client) GetKline(req global.KlineReq) ([]global.Kline, error) {

// 	period := req.Period
// 	if strings.Contains(period, "m") {
// 		period = period + "in"
// 	} else if period == "1h" {
// 		period = "60m"
// 	} else if strings.Contains(period, "h") {
// 		period = period + "our"
// 	} else if strings.Contains(period, "d") {
// 		period = period + "ay"
// 	} else if strings.Contains(period, "w") {
// 		period = period + "eek"
// 	}
// 	arg := make(map[string]string)
// 	arg["symbol"] = strings.ToLower(req.Base + req.Quote)
// 	if req.Count != 0 {
// 		arg["size"] = strconv.FormatInt(req.Count, 10)
// 	}
// 	arg["period"] = period

// 	r := struct {
// 		Status string  `json:"status"`
// 		Ch     string  `json:"ch"`
// 		Data   []Kline `json:"data"`
// 		Errmsg string  `json:"err-msg"`
// 	}{}
// 	e := c.doHTTP("GET", "/market/history/kline", arg, &r)
// 	if e != nil {
// 		return nil, e
// 	}
// 	if r.Status != "ok" {
// 		return nil, fmt.Errorf(r.Errmsg)
// 	}
// 	ik := []global.Kline{}
// 	for _, k := range r.Data {
// 		ik = append(ik, global.Kline{
// 			Base:      k.Base,
// 			Quote:     k.Quote,
// 			Timestamp: int64(k.Timestamp),
// 			High:      k.High,
// 			Open:      k.Open,
// 			Low:       k.Low,
// 			Close:     k.Close,
// 			Volume:    k.Volume,
// 		})
// 	}
// 	return ik, nil
// }

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
