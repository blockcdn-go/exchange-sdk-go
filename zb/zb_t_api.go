package zb

import (
	"errors"
	"fmt"
	"strings"

	"github.com/blockcdn-go/exchange-sdk-go/global"
	"github.com/blockcdn-go/exchange-sdk-go/utils"
)

// GetFund 获取帐号资金余额
func (c *Client) GetFund(global.FundReq) ([]global.Fund, error) {
	arg := map[string]interface{}{}
	arg["method"] = "getAccountInfo"

	f := []map[string]interface{}{}
	r := struct {
		errInfo
		Result struct {
			Coins interface{} `json:"coins"`
		} `json:"result"`
	}{}
	r.Result.Coins = &f
	err := c.httpReq("GET", "https://trade.zb.com/api/getAccountInfo", arg, &r, true)
	if err != nil {
		return nil, err
	}
	if r.errInfo.Code != 0 {
		return nil, errors.New(r.errInfo.Message)
	}
	ret := []global.Fund{}
	for _, co := range f {
		ret = append(ret, global.Fund{
			Base:      utils.ToString(co["key"]),
			Frozen:    utils.ToFloat(co["freez"]),
			Available: utils.ToFloat(co["available"]),
		})
	}
	return ret, nil
}

// InsertOrder 下单
func (c *Client) InsertOrder(req global.InsertReq) (global.InsertRsp, error) {
	arg := map[string]interface{}{}
	arg["method"] = "order"
	arg["price"] = req.Price
	arg["amount"] = req.Num
	arg["tradeType"] = utils.Ternary(req.Direction == 0, 1, 0)
	arg["acctType"] = 0
	arg["currency"] = strings.ToLower(req.Base + "_" + req.Quote)

	r := struct {
		errInfo
		ID string `json:"id"`
	}{}
	err := c.httpReq("GET", "https://trade.zb.com/api/order", arg, &r, true)
	if err != nil {
		return global.InsertRsp{}, err
	}
	if r.errInfo.Code != 0 {
		return global.InsertRsp{}, errors.New(r.errInfo.Message)
	}

	return global.InsertRsp{OrderNo: r.ID}, nil
}

// CancelOrder 撤销一个订单请求
// 注意，返回OK表示撤单请求成功。订单是否撤销成功请调用订单查询接口查询该订单状态
func (c *Client) CancelOrder(req global.CancelReq) error {
	arg := map[string]interface{}{}
	arg["method"] = "cancelOrder"
	arg["id"] = req.OrderNo
	arg["currency"] = strings.ToLower(req.Base + "_" + req.Quote)
	r := errInfo{}
	err := c.httpReq("GET", "https://trade.zb.com/api/cancelOrder", arg, &r, true)
	if err != nil {
		return err
	}
	if r.Code != 0 {
		return errors.New(r.Message)
	}
	return nil
}

// OrderStatus 查询某个订单详情
func (c *Client) OrderStatus(req global.StatusReq) (global.StatusRsp, error) {
	ret := global.StatusRsp{}
	arg := map[string]interface{}{}
	arg["method"] = "getOrder"
	arg["id"] = req.OrderNo
	arg["currency"] = strings.ToLower(req.Base + "_" + req.Quote)

	r := map[string]interface{}{}
	err := c.httpReq("GET", "https://trade.zb.com/api/getOrder", arg, &r, true)
	if err != nil {
		return ret, err
	}
	if cd, ok := r["code"]; ok && int(utils.ToFloat(cd)) != 0 {
		return ret, errors.New(utils.ToString(r["message"]))
	}

	// status : 挂单状态(1：取消,2：交易完成,0/3：待成交/待成交未交易部份)
	ret.TradeNum = utils.ToFloat(r["trade_amount"])
	ret.TradePrice = utils.ToFloat(r["price"])
	status := int(utils.ToFloat(r["status"]))
	if status == 1 {
		ret.Status = global.CANCELED
		ret.StatusMsg = "已撤单"
	} else if status == 2 {
		ret.Status = global.COMPLETETRADE
		ret.StatusMsg = "完全成交"
	}
	fmt.Printf("zb order status %+v\n", r)
	return ret, nil
}
