package weex

import (
	"errors"
	"fmt"
	"strings"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

var deftSymbol = []global.TradeSymbol{
	{Base: "BTC", Quote: "USD"},
	{Base: "BCH", Quote: "USD"},
	{Base: "LTC", Quote: "USD"},
	{Base: "ETH", Quote: "USD"},
	{Base: "ZEC", Quote: "USD"},
	{Base: "DASH", Quote: "USD"},
	{Base: "ETC", Quote: "USD"},
	{Base: "BCH", Quote: "BTC"},
	{Base: "LTC", Quote: "BTC"},
	{Base: "ETH", Quote: "BTC"},
	{Base: "ZEC", Quote: "BTC"},
	{Base: "DASH", Quote: "BTC"},
	{Base: "ETC", Quote: "BTC"},
}

type weexRsp struct {
	Code int         `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
}

func (c *Client) recursionOrderStatus(page int, req global.StatusReq) (global.StatusRsp, error) {
	in := map[string]interface{}{}
	in["access_id"] = req.APIKey
	in["page"] = page
	in["market"] = strings.ToUpper(req.Base + req.Quote)
	in["limit"] = 100

	data := struct {
		HasNext bool                     `json:"has_next"`
		Data    []map[string]interface{} `json:"data"`
	}{}
	r := weexRsp{Data: &data}
	err := c.httpReq("GET", "https://api.weex.com/v1/order/pending", in, &r, true)
	if err != nil {
		return global.StatusRsp{}, err
	}
	if r.Code != 0 {
		return global.StatusRsp{}, errors.New(r.Msg)
	}
	fmt.Printf("weex orderstatus %+v\n", r)

	// 遍历data 找到订单号和请求订单号相同的订单
	for _, d := range data.Data {
		if toString(d["id"]) == req.OrderNo {
			return global.StatusRsp{
				TradeNum:   toFloat(d["deal_amount"]),
				TradePrice: toFloat(d["avg_price"]),
				Status:     global.COMPLETETRADE,
				StatusMsg:  "完全成交",
			}, nil
		}
	}
	// 如果有下一条数据则查询下一条数据
	if data.HasNext {
		return c.recursionOrderStatus(page+1, req)
	}
	// 已经没有下一条数据了，当前订单应该还没有成交，返回一条空数据
	return global.StatusRsp{}, nil
}
