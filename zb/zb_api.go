package zb

import (
	"errors"
	"fmt"
	"strings"

	"github.com/blockcdn-go/exchange-sdk-go/utils"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

type errInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// GetAllSymbol 交易市场详细行情接口
func (c *Client) GetAllSymbol() ([]global.TradeSymbol, error) {
	r := map[string]interface{}{}
	err := c.httpReq("GET", "http://api.zb.com/data/v1/markets", nil, &r, false)
	ret := []global.TradeSymbol{}
	for k := range r {
		base, quote := split3(k)
		ret = append(ret, global.TradeSymbol{
			Base:  base,
			Quote: quote,
		})
	}
	return ret, err
}

// GetDepth 获取深度行情
func (c *Client) GetDepth(req global.TradeSymbol) (global.Depth, error) {

	arg := map[string]interface{}{}
	arg["market"] = strings.ToLower(req.Base + "_" + req.Quote)
	arg["size"] = 100

	r := struct {
		errInfo
		Asks [][]float64 `json:"asks"`
		Bids [][]float64 `json:"bids"`
	}{}
	err := c.httpReq("GET", "http://api.zb.com/data/v1/depth", arg, &r, false)
	if err != nil {
		return global.Depth{}, err
	}
	if r.errInfo.Code != 0 {
		return global.Depth{}, errors.New(r.errInfo.Message)
	}
	dp := global.Depth{
		Base:  req.Base,
		Quote: req.Quote,
		Asks:  []global.DepthPair{},
		Bids:  []global.DepthPair{},
	}
	for _, a := range r.Asks {
		if len(a) < 2 {
			continue
		}
		dp.Asks = append(dp.Asks, global.DepthPair{Price: a[0], Size: a[1]})
	}
	for _, b := range r.Bids {
		if len(b) < 2 {
			continue
		}
		dp.Bids = append(dp.Bids, global.DepthPair{Price: b[0], Size: b[1]})
	}
	return dp, nil
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
	arg := map[string]interface{}{}
	arg["market"] = strings.ToLower(req.Base + "_" + req.Quote)
	arg["type"] = period
	arg["size"] = utils.Ternary(req.Count == 0, 500, req.Count)
	r := struct {
		errInfo
		Data [][]float64 `json:"data"`
	}{}
	err := c.httpReq("GET", "http://api.zb.com/data/v1/kline", arg, &r, false)
	if err != nil {
		return nil, err
	}
	if r.errInfo.Code != 0 {
		return nil, errors.New(r.errInfo.Message)
	}
	kline := []global.Kline{}
	for _, k1 := range r.Data {
		if len(k1) < 6 {
			continue
		}
		kline = append(kline, global.Kline{
			Base:      req.Base,
			Quote:     req.Quote,
			Timestamp: int64(k1[0]),
			Open:      k1[1],
			High:      k1[2],
			Low:       k1[3],
			Close:     k1[4],
			Volume:    k1[5],
		})
	}
	return kline, nil
}

/////////////////////////////////////////////////////////////////////////

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
