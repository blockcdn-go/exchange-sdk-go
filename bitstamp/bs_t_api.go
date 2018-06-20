package bitstamp

import (
	"errors"
	"strings"

	"github.com/blockcdn-go/exchange-sdk-go/utils"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

// GetFund 获取帐号资金余额
func (c *Client) GetFund(global.FundReq) ([]global.Fund, error) {
	in := map[string]interface{}{}
	r := map[string]interface{}{}
	err := c.httpReq("POST", "https://www.bitstamp.net/api/v2/balance/", in, &r, true)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// InsertOrder 下单交易
func (c *Client) InsertOrder(req global.InsertReq) (global.InsertRsp, error) {
	in := map[string]interface{}{}
	in["amount"] = req.Num

	path := "https://www.bitstamp.net/api/v2/"
	if req.Direction == 0 {
		path += "buy/"
	} else {
		path += "sell/"
	}
	if req.Type == 1 {
		path += "market/"
	} else {
		in["price"] = req.Price
	}
	path += strings.ToLower(req.Base + "_" + req.Quote + "/")

	r := map[string]interface{}{}
	err := c.httpReq("POST", path, in, &r, true)
	if err != nil {
		return global.InsertRsp{}, err
	}
	if r["status"] == "error" {
		return global.InsertRsp{}, errors.New(utils.ToString(r["reason"]))
	}
	return global.InsertRsp{OrderNo: utils.ToString(r["id"])}, nil
}

// CancelOrder 取消订单
// 通过测试，第一个参数对结果没有影响，只要orderno正确就能取消订单，
// 但是如果第一个参数填入错误的代码将返回错误，但是订单依然被取消了
func (c *Client) CancelOrder(req global.CancelReq) error {
	in := map[string]interface{}{}
	in["id"] = req.OrderNo
	r := map[string]interface{}{}
	err := c.httpReq("POST", "https://www.bitstamp.net/api/v2/cancel_order/", in, &r, true)
	if err != nil {
		return err
	}
	if r["status"] == "error" {
		return errors.New(utils.ToString(r["reason"]))
	}
	return nil
}

// OrderStatus 获取订单状态
func (c *Client) OrderStatus(req global.StatusReq) (global.StatusRsp, error) {
	in := map[string]interface{}{}
	r := map[string]interface{}{}
	err := c.httpReq("POST", "https://www.bitstamp.net/api/order_status/", in, &r, true)
	if err != nil {
		return global.StatusRsp{}, err
	}
	if utils.ToString(r["status"]) == "error" {
		return global.StatusRsp{}, errors.New(utils.ToString(r["reason"]))
	}
	status := utils.ToString(r["status"])
	m := global.StatusRsp{}
	if status == "Finished" {
		m.Status = global.COMPLETETRADE
		m.StatusMsg = "完全成交"
	}
	return m, nil
}
