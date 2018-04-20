package huobi

import (
	"fmt"
)

// Account ...
type Account struct {
	AccountID   int64  `json:"id"`
	Status      string `json:"state"` // working：正常, lock：账户被锁定
	AccountType string `json:"type"`  // spot：现货账户
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

// Balance ...
type Balance struct {
	Amount     string `json:"balance"`  // 余额
	CurrencyNo string `json:"currency"` //币种
	BType      string `json:"type"`     // trade: 交易余额，frozen: 冻结余额
}

// BalanceInfo 查询指定账户的余额
// @parm ishadax true 从HADAX站查询, false pro 站查询
// @parm accountID GetAllAccountID函数返回的id
func (c *Client) BalanceInfo(ishadax bool, accountID int64) ([]Balance, error) {
	f := func() string {
		if ishadax {
			return "/v1/hadax/account/accounts/"
		}
		return "/v1/account/accounts/"
	}

	r := struct {
		Status string `json:"status"`
		Data   struct {
			List []Balance `json:"list"`
		} `json:"data"`

		Errmsg string `json:"err-msg"`
	}{}

	path := fmt.Sprintf("%s%d/balance", f(), accountID)
	e := c.doHTTP("GET", path, nil, &r)
	if e != nil {
		return nil, e
	}
	if r.Status != "ok" {
		return nil, fmt.Errorf(r.Errmsg)
	}
	return r.Data.List, nil
}

// InsertOrderReq ...
type InsertOrderReq struct {
	AccountID string `json:"account-id"` // 账户ID
	Amount    string `json:"amount"`     // 限价表示下单数量, 市价买单时表示买多少钱, 市价卖单时表示卖多少币
	Price     string `json:"price"`      // 下单价格, 市价单不传该参数
	Source    string `json:"source"`     // 订单来源, api: API调用, margin-api: 借贷资产交易
	Symbol    string `json:"symbol"`     // 交易对, btcusdt, bccbtc......
	Type      string `json:"type"`       // 订单类型, buy-market: 市价买, sell-market: 市价卖, buy-limit: 限价买, sell-limit: 限价卖
}

// InsertOrder 下单
// @return string: orderNo
func (c *Client) InsertOrder(ishadax bool, req InsertOrderReq) (string, error) {
	req.Source = "api"
	mapParams := if2map(req)

	f := func() string {
		if ishadax {
			return "/v1/hadax/order/orders/place"
		}
		return "/v1/order/orders/place"
	}
	r := struct {
		Status string `json:"status"`
		Errmsg string `json:"err-msg"`
		Data   string `json:"data"`
	}{}
	e := c.doHTTP("POST", f(), mapParams, &r)
	if e != nil {
		return "", e
	}
	if r.Status != "ok" {
		return "", fmt.Errorf(r.Errmsg)
	}
	return r.Data, nil
}
