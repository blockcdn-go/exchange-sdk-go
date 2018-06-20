package weex

import (
	"errors"
	"fmt"
	"strings"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

// GetFund 获取帐号资金余额
func (c *Client) GetFund(global.FundReq) ([]global.Fund, error) {
	d := map[string]struct {
		Available string `json:"available"`
		Frozen    string `json:"frozen"`
	}{}
	r := weexRsp{Data: &d}
	err := c.httpReq("GET", "https://api.weex.com/v1/balance/", nil, &r, true)
	if err != nil {
		return nil, err
	}
	if r.Code != 0 {
		return nil, errors.New(r.Msg)
	}
	ret := []global.Fund{}
	for k, v := range d {
		ret = append(ret, global.Fund{
			Base:      k,
			Available: toFloat(v.Available),
			Frozen:    toFloat(v.Frozen),
		})
	}
	return ret, nil
}

// InsertOrder 下单
func (c *Client) InsertOrder(req global.InsertReq) (global.InsertRsp, error) {
	t := "market"
	d := "buy"
	in := map[string]interface{}{}
	if req.Type == 0 {
		t = "limit"
		in["price"] = toString(req.Price)
	}
	if req.Direction == 1 {
		d = "sell"
	}
	path := fmt.Sprintf("https://api.weex.com/v1/order/%s", t)

	in["access_id"] = req.APIKey
	in["market"] = strings.ToUpper(req.Base + req.Quote)
	in["type"] = d
	in["amount"] = toString(req.Num)

	data := map[string]interface{}{}
	r := weexRsp{Data: &data}
	err := c.httpReq("POST", path, in, &r, true)
	if err != nil {
		return global.InsertRsp{}, err
	}

	if r.Code != 0 {
		return global.InsertRsp{}, errors.New(r.Msg)
	}
	fmt.Printf("weex insert %+v\n", r)
	return global.InsertRsp{OrderNo: toString(data["id"])}, nil
}

// CancelOrder 撤销一个订单请求
// 注意，返回OK表示撤单请求成功。订单是否撤销成功请调用订单查询接口查询该订单状态
func (c *Client) CancelOrder(req global.CancelReq) error {
	in := map[string]interface{}{}
	in["access_id"] = req.APIKey
	in["order_id"] = int64(toFloat(req.OrderNo))
	in["market"] = strings.ToUpper(req.Base + req.Quote)

	data := map[string]interface{}{}
	r := weexRsp{Data: &data}
	err := c.httpReq("DELETE", "https://api.weex.com/v1/order/pending", in, &r, true)
	if err != nil {
		return err
	}
	if r.Code != 0 {
		return errors.New(r.Msg)
	}
	fmt.Printf("weex cancel %+v\n", r)
	return nil
}

// OrderStatus 查询某个订单详情
// @note: api不能根据订单号进行查询，所有智能通过查询成交再根据订单号进行筛选的方式判断
func (c *Client) OrderStatus(req global.StatusReq) (global.StatusRsp, error) {
	return c.recursionOrderStatus(1, req)
}
