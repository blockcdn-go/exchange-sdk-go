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

// GetAllSymbol 交易市场详细行情接口
func (c *Client) GetAllSymbol() ([]global.TradeSymbol, error) {
	d := []struct {
		Quote string `json:"buy_asset_type"`
		Base  string `json:"sell_asset_type"`
	}{}
	r := weexRsp{Data: &d}
	err := c.httpReq("GET", "https://www.weexpro.com/exchange/v1/market/info", nil, &r, false)
	if err != nil {
		return nil, err
	}
	if r.Code != 0 {
		return nil, errors.New(r.Msg)
	}
	ret := []global.TradeSymbol{}
	for _, s := range d {
		ret = append(ret, global.TradeSymbol{
			Base:  s.Base,
			Quote: s.Quote,
		})
	}
	return ret, nil
}

// GetKline 获取k线数据
func (c *Client) GetKline(req global.KlineReq) ([]global.Kline, error) {
	period := req.Period
	if strings.Contains(period, "m") {
		period = period + "in"
	} else if strings.Contains(period, "h") {
		period = period + "our"
	} else if strings.Contains(period, "d") {
		period = period + "ay"
	} else if strings.Contains(period, "w") {
		period = period + "eek"
	}
	sybmol := strings.ToLower(req.Base + req.Quote)
	in := make(map[string]interface{})
	in["market"] = sybmol
	in["type"] = period
	d := [][]interface{}{}
	r := weexRsp{Data: &d}
	err := c.httpReq("GET", "https://api.weex.com/v1/market/kline", in, &r, false)
	if err != nil {
		return nil, err
	}
	if r.Code != 0 {
		return nil, errors.New(r.Msg)
	}
	ret := []global.Kline{}
	for _, v := range d {
		if len(v) < 6 {
			continue
		}
		ret = append(ret, global.Kline{
			Base:      req.Base,
			Quote:     req.Quote,
			Timestamp: int64(toFloat(v[0])) * 1000,
			Open:      toFloat(v[1]),
			Close:     toFloat(v[2]),
			High:      toFloat(v[3]),
			Low:       toFloat(v[4]),
			Volume:    toFloat(v[5]),
		})
	}
	return ret, nil
}

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
