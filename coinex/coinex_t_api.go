package coinex

import (
	"errors"
	"strings"

	"github.com/blockcdn-go/exchange-sdk-go/utils"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

// GetFund 获取帐号资金余额
func (c *Client) GetFund(global.FundReq) ([]global.Fund, error) {
	in := map[string]interface{}{}
	data := map[string]map[string]interface{}{}
	r := plainRsp{Data: &data}
	err := c.httpReq("GET", "https://api.coinex.com/v1/balance/", in, &r, true)
	if err != nil {
		return nil, err
	}
	if r.Code != 0 {
		return nil, errors.New(r.Message)
	}
	ret := []global.Fund{}
	for c, v := range data {
		cc := global.Fund{
			Base:      c,
			Available: utils.ToFloat(v["available"]),
			Frozen:    utils.ToFloat(v["frozen"]),
		}
		ret = append(ret, cc)
	}
	return ret, nil
}

// InsertOrder 下单接口
func (c *Client) InsertOrder(req global.InsertReq) (global.InsertRsp, error) {
	path := "https://api.coinex.com/v1/order/"
	in := map[string]interface{}{}
	data := map[string]interface{}{}
	r := plainRsp{Data: &data}
	in["market"] = strings.ToUpper(req.Base + req.Quote)
	in["amount"] = utils.ToString(req.Num)
	in["type"] = utils.Ternary(req.Direction == 0, "buy", "sell")
	if req.Type == 0 {
		path += "limit"
		in["price"] = utils.ToString(req.Price)
	} else {
		path += "market"
	}
	err := c.httpReq("POST", path, in, &r, true)
	if err != nil {
		return global.InsertRsp{}, err
	}
	if r.Code != 0 {
		return global.InsertRsp{}, errors.New(r.Message)
	}
	return global.InsertRsp{OrderNo: utils.ToString(data["id"])}, nil
}

// CancelOrder 撤单
func (c *Client) CancelOrder(req global.CancelReq) error {
	in := map[string]interface{}{}
	data := map[string]interface{}{}
	r := plainRsp{Data: &data}
	in["market"] = strings.ToUpper(req.Base + req.Quote)
	in["id"] = int(utils.ToFloat(req.OrderNo))
	err := c.httpReq("POST", "https://api.coinex.com/v1/order/pending", in, &r, true)
	if err != nil {
		return err
	}
	if r.Code != 0 {
		return errors.New(r.Message)
	}
	return nil
}

// OrderStatus 获取订单状态
func (c *Client) OrderStatus(req global.StatusReq) (global.StatusRsp, error) {
	in := map[string]interface{}{}
	data := map[string]interface{}{}
	r := plainRsp{Data: &data}
	in["market"] = strings.ToUpper(req.Base + req.Quote)
	in["id"] = int(utils.ToFloat(req.OrderNo))
	err := c.httpReq("POST", "https://api.coinex.com/v1/order/", in, &r, true)
	if err != nil {
		return global.StatusRsp{}, err
	}
	if r.Code != 0 {
		return global.StatusRsp{}, errors.New(r.Message)
	}
	ret := global.StatusRsp{}
	status := utils.ToString(data["status"])
	ret.TradeNum = utils.ToFloat(data["deal_amount"])
	if ret.TradeNum != 0. {
		ret.TradePrice = utils.ToFloat(data["deal_money"]) / ret.TradeNum
	}
	if status == "done" {
		ret.Status = global.COMPLETETRADE
		ret.StatusMsg = "完全成交"
	} else if status == "not_deal" || status == "part_deal" {
		// empty
	} else {
		ret.Status = global.CANCELED
		ret.StatusMsg = "已撤单"
	}
	return ret, nil
}
