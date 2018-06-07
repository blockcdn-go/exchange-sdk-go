package binance

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitlab.mybcdn.com/golang/blockcoin/apidb"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

type rawExecutedOrder struct {
	Symbol           string  `json:"symbol"`
	OrderID          int     `json:"orderId"`
	ClientOrderID    string  `json:"clientOrderId"`
	Price            string  `json:"price"`
	OrigQty          string  `json:"origQty"`
	ExecutedQuoteQty string  `json:"executedQuoteQty"`
	ExecutedQty      string  `json:"executedQty"`
	Status           string  `json:"status"`
	TimeInForce      string  `json:"timeInForce"`
	Type             string  `json:"type"`
	Side             string  `json:"side"`
	StopPrice        string  `json:"stopPrice"`
	IcebergQty       string  `json:"icebergQty"`
	Time             float64 `json:"time"`
}

func (as *apiService) InsertOrder(or global.InsertReq) (global.InsertRsp, error) {
	params := make(map[string]string)
	params["symbol"] = strings.ToUpper(or.Base + or.Quote)
	params["side"] = string(SideBuy)
	if or.Direction == 1 {
		params["side"] = string(SideSell)
	}
	params["type"] = string(TypeLimit)
	if or.Type == 1 {
		params["type"] = string(TypeMarket)
	}
	if or.Type == int(apidb.LIMIT) {
		// 限价才有的参数
		params["timeInForce"] = string(GTC)
		params["price"] = strconv.FormatFloat(or.Price, 'f', -1, 64)
	}
	params["quantity"] = strconv.FormatFloat(or.Num, 'f', -1, 64)
	params["timestamp"] = strconv.FormatInt(time.Now().Unix()*1000, 10)
	// if or.NewClientOrderID != "" {
	// 	params["newClientOrderId"] = or.NewClientOrderID
	// }
	// if or.StopPrice != 0 {
	// 	params["stopPrice"] = strconv.FormatFloat(or.StopPrice, 'f', -1, 64)
	// }
	// if or.IcebergQty != 0 {
	// 	params["icebergQty"] = strconv.FormatFloat(or.IcebergQty, 'f', -1, 64)
	// }
	rawOrder := struct {
		Symbol        string  `json:"symbol"`
		OrderID       int64   `json:"orderId"`
		ClientOrderID string  `json:"clientOrderId"`
		TransactTime  float64 `json:"transactTime"`
	}{}
	err := as.request("POST", "api/v3/order", params, &rawOrder, true, true)
	if err != nil {
		return global.InsertRsp{}, err
	}
	return global.InsertRsp{
		OrderNo: strconv.FormatInt(rawOrder.OrderID, 10),
	}, nil

	// t, err := timeFromUnixTimestampFloat(rawOrder.TransactTime)
	// if err != nil {
	// 	return global.InsertRsp{}, err
	// }

	// return &ProcessedOrder{
	// 	Symbol:        rawOrder.Symbol,
	// 	OrderID:       rawOrder.OrderID,
	// 	ClientOrderID: rawOrder.ClientOrderID,
	// 	TransactTime:  t,
	// }, nil
}

// func (as *apiService) OrderStatus(qor global.StatusReq) (global.StatusRsp, error) {
// 	params := make(map[string]string)
// 	params["symbol"] = strings.ToUpper(qor.Base + qor.Quote)
// 	params["timestamp"] = strconv.FormatInt(time.Now().Unix()*1000, 10)
// 	if qor.OrderNo != "" {
// 		params["orderId"] = qor.OrderNo
// 	}
// 	// if qor.OrigClientOrderID != "" {
// 	// 	params["origClientOrderId"] = qor.OrigClientOrderID
// 	// }
// 	params["recvWindow"] = strconv.FormatInt(recvWindow(time.Second*5), 10)
// 	rawOrder := &rawExecutedOrder{}
// 	err := as.request("GET", "api/v3/order", params, rawOrder, true, true)
// 	if err != nil {
// 		return global.StatusRsp{}, err
// 	}

// 	or, err := executedOrderFromRaw(rawOrder)
// 	if err != nil {
// 		return global.StatusRsp{}, err
// 	}
// 	m := global.StatusRsp{}
// 	m.TradePrice = or.ExecutePrice
// 	m.TradeNum = or.ExecutedQty

// 	if or.ExecutedQty != 0. || or.Status == StatusPartiallyFilled {
// 		m.Status = global.HALFTRADE
// 		m.StatusMsg = "部分成交"
// 	}
// 	if or.ExecutedQty == or.OrigQty || or.Status == StatusFilled {
// 		m.Status = global.COMPLETETRADE
// 		m.StatusMsg = "完全成交"
// 	}
// 	if or.Status == StatusCancelled {
// 		m.Status = global.CANCELED
// 		m.StatusMsg = "已撤单"
// 	}
// 	if or.Status == StatusRejected {
// 		m.Status = global.FAILED
// 		m.StatusMsg = "订单被拒绝"
// 	}
// 	if or.Status == StatusExpired {
// 		m.Status = global.FAILED
// 		m.StatusMsg = "订单超时"
// 	}
// 	fmt.Printf("binance order status %+v\n", or)
// 	return m, nil
// }

func (as *apiService) CancelOrder(cor global.CancelReq) error {
	params := make(map[string]string)
	params["symbol"] = strings.ToUpper(cor.Base + cor.Quote)
	params["timestamp"] = strconv.FormatInt(unixMillis(time.Now()), 10)
	if cor.OrderNo != "" {
		params["orderId"] = cor.OrderNo
	}
	// if cor.OrigClientOrderID != "" {
	// 	params["origClientOrderId"] = cor.OrigClientOrderID
	// }
	// if cor.NewClientOrderID != "" {
	// 	params["newClientOrderId"] = cor.NewClientOrderID
	// }
	params["recvWindow"] = strconv.FormatInt(recvWindow(time.Second*5), 10)

	rawCanceledOrder := struct {
		Symbol            string `json:"symbol"`
		OrigClientOrderID string `json:"origClientOrderId"`
		OrderID           int64  `json:"orderId"`
		ClientOrderID     string `json:"clientOrderId"`
	}{}
	err := as.request("DELETE", "api/v3/order", params, &rawCanceledOrder, true, true)
	return err
	// if err != nil {
	// 	return err
	// }
	// return &CanceledOrder{
	// 	Symbol:            rawCanceledOrder.Symbol,
	// 	OrigClientOrderID: rawCanceledOrder.OrigClientOrderID,
	// 	OrderID:           rawCanceledOrder.OrderID,
	// 	ClientOrderID:     rawCanceledOrder.ClientOrderID,
	// }, nil
}

func (as *apiService) OpenOrders(oor OpenOrdersRequest) ([]*ExecutedOrder, error) {
	params := make(map[string]string)
	params["symbol"] = oor.Symbol
	params["timestamp"] = strconv.FormatInt(unixMillis(oor.Timestamp), 10)
	if oor.RecvWindow != 0 {
		params["recvWindow"] = strconv.FormatInt(recvWindow(oor.RecvWindow), 10)
	}
	rawOrders := []*rawExecutedOrder{}
	err := as.request("GET", "api/v3/openOrders", params, &rawOrders, true, true)
	if err != nil {
		return nil, err
	}
	var eoc []*ExecutedOrder
	for _, rawOrder := range rawOrders {
		eo, err := executedOrderFromRaw(rawOrder)
		if err != nil {
			return nil, err
		}
		eoc = append(eoc, eo)
	}

	return eoc, nil
}

func (as *apiService) OrderStatus(qor global.StatusReq) (global.StatusRsp, error) {
	params := make(map[string]string)
	params["symbol"] = strings.ToUpper(qor.Base + qor.Quote)
	params["timestamp"] = strconv.FormatInt(unixMillis(time.Now()), 10)
	params["orderId"] = qor.OrderNo
	params["recvWindow"] = strconv.FormatInt(recvWindow(time.Second*5), 10)

	// if aor.Limit != 0 {
	// 	params["limit"] = strconv.Itoa(aor.Limit)
	// }

	rawOrders := []rawExecutedOrder{}
	err := as.request("GET", "api/v3/allOrders", params, &rawOrders, true, true)
	if err != nil {
		return global.StatusRsp{}, err
	}

	var eoc []*ExecutedOrder
	for _, rawOrder := range rawOrders {
		eo, err := executedOrderFromRaw(&rawOrder)
		if err != nil {
			return global.StatusRsp{}, err
		}
		eoc = append(eoc, eo)
	}
	if len(eoc) != 1 {
		return global.StatusRsp{}, fmt.Errorf("rsp len error: %d", len(eoc))
	}
	or := eoc[0]
	m := global.StatusRsp{}
	m.TradePrice = or.ExecutePrice
	m.TradeNum = or.ExecutedQty

	if or.Status == StatusPartiallyFilled {
		m.Status = global.HALFTRADE
		m.StatusMsg = "部分成交"
	}
	if or.Status == StatusFilled {
		m.Status = global.COMPLETETRADE
		m.StatusMsg = "完全成交"
	}
	if or.Status == StatusCancelled {
		m.Status = global.CANCELED
		m.StatusMsg = "已撤单"
	}
	if or.Status == StatusRejected {
		m.Status = global.FAILED
		m.StatusMsg = "订单被拒绝"
	}
	if or.Status == StatusExpired {
		m.Status = global.FAILED
		m.StatusMsg = "订单超时"
	}
	fmt.Printf("binance order status %+v\n", or)
	return m, nil
}

func (as *apiService) GetFund(global.FundReq) ([]global.Fund, error) {
	params := make(map[string]string)
	params["timestamp"] = strconv.FormatInt(unixMillis(time.Now()), 10)
	params["recvWindow"] = strconv.FormatInt(recvWindow(5*time.Second), 10)

	rawAccount := struct {
		MakerCommision   int64 `json:"makerCommision"`
		TakerCommission  int64 `json:"takerCommission"`
		BuyerCommission  int64 `json:"buyerCommission"`
		SellerCommission int64 `json:"sellerCommission"`
		CanTrade         bool  `json:"canTrade"`
		CanWithdraw      bool  `json:"canWithdraw"`
		CanDeposit       bool  `json:"canDeposit"`
		Balances         []struct {
			Asset  string `json:"asset"`
			Free   string `json:"free"`
			Locked string `json:"locked"`
		}
	}{}
	err := as.request("GET", "api/v3/account", params, &rawAccount, true, true)
	if err != nil {
		return nil, err
	}

	// acc := &Account{
	// 	MakerCommision:  rawAccount.MakerCommision,
	// 	TakerCommision:  rawAccount.TakerCommission,
	// 	BuyerCommision:  rawAccount.BuyerCommission,
	// 	SellerCommision: rawAccount.SellerCommission,
	// 	CanTrade:        rawAccount.CanTrade,
	// 	CanWithdraw:     rawAccount.CanWithdraw,
	// 	CanDeposit:      rawAccount.CanDeposit,
	// }

	ar := []global.Fund{}
	for _, b := range rawAccount.Balances {
		f, err := floatFromString(b.Free)
		if err != nil {
			return nil, err
		}
		l, err := floatFromString(b.Locked)
		if err != nil {
			return nil, err
		}
		// acc.Balances = append(acc.Balances, &Balance{
		// 	Asset:  b.Asset,
		// 	Free:   f,
		// 	Locked: l,
		// })
		ar = append(ar, global.Fund{
			Base:      b.Asset,
			Available: f,
			Frozen:    l,
		})
	}

	return ar, nil
}

func (as *apiService) MyTrades(mtr MyTradesRequest) ([]*Trade, error) {
	params := make(map[string]string)
	params["symbol"] = mtr.Symbol
	params["timestamp"] = strconv.FormatInt(unixMillis(mtr.Timestamp), 10)
	if mtr.RecvWindow != 0 {
		params["recvWindow"] = strconv.FormatInt(recvWindow(mtr.RecvWindow), 10)
	}
	if mtr.FromID != 0 {
		params["orderId"] = strconv.FormatInt(mtr.FromID, 10)
	}
	if mtr.Limit != 0 {
		params["limit"] = strconv.Itoa(mtr.Limit)
	}
	rawTrades := []struct {
		ID              int64   `json:"id"`
		Price           string  `json:"price"`
		Qty             string  `json:"qty"`
		Commission      string  `json:"commission"`
		CommissionAsset string  `json:"commissionAsset"`
		Time            float64 `json:"time"`
		IsBuyer         bool    `json:"isBuyer"`
		IsMaker         bool    `json:"isMaker"`
		IsBestMatch     bool    `json:"isBestMatch"`
	}{}
	err := as.request("GET", "api/v3/myTrades", params, &rawTrades, true, true)
	if err != nil {
		return nil, err
	}

	var tc []*Trade
	for _, rt := range rawTrades {
		price, err := floatFromString(rt.Price)
		if err != nil {
			return nil, err
		}
		qty, err := floatFromString(rt.Qty)
		if err != nil {
			return nil, err
		}
		commission, err := floatFromString(rt.Commission)
		if err != nil {
			return nil, err
		}
		t, err := timeFromUnixTimestampFloat(rt.Time)
		if err != nil {
			return nil, err
		}
		tc = append(tc, &Trade{
			ID:              rt.ID,
			Price:           price,
			Qty:             qty,
			Commission:      commission,
			CommissionAsset: rt.CommissionAsset,
			Time:            t,
			IsBuyer:         rt.IsBuyer,
			IsMaker:         rt.IsMaker,
			IsBestMatch:     rt.IsBestMatch,
		})
	}
	return tc, nil
}

func (as *apiService) Withdraw(wr WithdrawRequest) (*WithdrawResult, error) {
	params := make(map[string]string)
	params["asset"] = wr.Asset
	params["address"] = wr.Address
	params["amount"] = strconv.FormatFloat(wr.Amount, 'f', 10, 64)
	params["timestamp"] = strconv.FormatInt(unixMillis(wr.Timestamp), 10)
	if wr.RecvWindow != 0 {
		params["recvWindow"] = strconv.FormatInt(recvWindow(wr.RecvWindow), 10)
	}
	if wr.Name != "" {
		params["name"] = wr.Name
	}
	rawResult := struct {
		Msg     string `json:"msg"`
		Success bool   `json:"success"`
	}{}
	err := as.request("POST", "wapi/v1/withdraw.html", params, &rawResult, true, true)
	if err != nil {
		return nil, err
	}

	return &WithdrawResult{
		Msg:     rawResult.Msg,
		Success: rawResult.Success,
	}, nil
}
func (as *apiService) DepositHistory(hr HistoryRequest) ([]*Deposit, error) {
	params := make(map[string]string)
	params["timestamp"] = strconv.FormatInt(unixMillis(hr.Timestamp), 10)
	if hr.Asset != "" {
		params["asset"] = hr.Asset
	}
	if hr.Status != nil {
		params["status"] = strconv.Itoa(*hr.Status)
	}
	if !hr.StartTime.IsZero() {
		params["startTime"] = strconv.FormatInt(unixMillis(hr.StartTime), 10)
	}
	if !hr.EndTime.IsZero() {
		params["startTime"] = strconv.FormatInt(unixMillis(hr.EndTime), 10)
	}
	if hr.RecvWindow != 0 {
		params["recvWindow"] = strconv.FormatInt(recvWindow(hr.RecvWindow), 10)
	}

	rawDepositHistory := struct {
		DepositList []struct {
			InsertTime float64 `json:"insertTime"`
			Amount     float64 `json:"amount"`
			Asset      string  `json:"asset"`
			Status     int     `json:"status"`
		}
		Success bool `json:"success"`
	}{}
	err := as.request("POST", "wapi/v1/getDepositHistory.html", params, &rawDepositHistory, true, true)
	if err != nil {
		return nil, err
	}

	var dc []*Deposit
	for _, d := range rawDepositHistory.DepositList {
		t, err := timeFromUnixTimestampFloat(d.InsertTime)
		if err != nil {
			return nil, err
		}
		dc = append(dc, &Deposit{
			InsertTime: t,
			Amount:     d.Amount,
			Asset:      d.Asset,
			Status:     d.Status,
		})
	}

	return dc, nil
}
func (as *apiService) WithdrawHistory(hr HistoryRequest) ([]*Withdrawal, error) {
	params := make(map[string]string)
	params["timestamp"] = strconv.FormatInt(unixMillis(hr.Timestamp), 10)
	if hr.Asset != "" {
		params["asset"] = hr.Asset
	}
	if hr.Status != nil {
		params["status"] = strconv.Itoa(*hr.Status)
	}
	if !hr.StartTime.IsZero() {
		params["startTime"] = strconv.FormatInt(unixMillis(hr.StartTime), 10)
	}
	if !hr.EndTime.IsZero() {
		params["startTime"] = strconv.FormatInt(unixMillis(hr.EndTime), 10)
	}
	if hr.RecvWindow != 0 {
		params["recvWindow"] = strconv.FormatInt(recvWindow(hr.RecvWindow), 10)
	}
	rawWithdrawHistory := struct {
		WithdrawList []struct {
			Amount    float64 `json:"amount"`
			Address   string  `json:"address"`
			TxID      string  `json:"txId"`
			Asset     string  `json:"asset"`
			ApplyTime float64 `json:"insertTime"`
			Status    int     `json:"status"`
		}
		Success bool `json:"success"`
	}{}
	err := as.request("POST", "wapi/v1/getWithdrawHistory.html", params, &rawWithdrawHistory, true, true)
	if err != nil {
		return nil, err
	}

	var wc []*Withdrawal
	for _, w := range rawWithdrawHistory.WithdrawList {
		t, err := timeFromUnixTimestampFloat(w.ApplyTime)
		if err != nil {
			return nil, err
		}
		wc = append(wc, &Withdrawal{
			Amount:    w.Amount,
			Address:   w.Address,
			TxID:      w.TxID,
			Asset:     w.Asset,
			ApplyTime: t,
			Status:    w.Status,
		})
	}

	return wc, nil
}

func executedOrderFromRaw(reo *rawExecutedOrder) (*ExecutedOrder, error) {
	price, err := strconv.ParseFloat(reo.Price, 64)
	if err != nil {
		return nil, warpError(err, "cannot parse Order.CloseTime")
	}
	origQty, err := strconv.ParseFloat(reo.OrigQty, 64)
	if err != nil {
		return nil, warpError(err, "cannot parse Order.OrigQty")
	}
	exePrice, _ := strconv.ParseFloat(reo.ExecutedQuoteQty, 64)
	execQty, err := strconv.ParseFloat(reo.ExecutedQty, 64)
	if err != nil {
		return nil, warpError(err, "cannot parse Order.ExecutedQty")
	}
	stopPrice, err := strconv.ParseFloat(reo.StopPrice, 64)
	if err != nil {
		return nil, warpError(err, "cannot parse Order.StopPrice")
	}
	icebergQty, err := strconv.ParseFloat(reo.IcebergQty, 64)
	if err != nil {
		return nil, warpError(err, "cannot parse Order.IcebergQty")
	}
	t, err := timeFromUnixTimestampFloat(reo.Time)
	if err != nil {
		return nil, warpError(err, "cannot parse Order.CloseTime")
	}

	return &ExecutedOrder{
		Symbol:        reo.Symbol,
		OrderID:       reo.OrderID,
		ClientOrderID: reo.ClientOrderID,
		Price:         price,
		OrigQty:       origQty,
		ExecutePrice:  exePrice,
		ExecutedQty:   execQty,
		Status:        OrderStatus(reo.Status),
		TimeInForce:   TimeInForce(reo.TimeInForce),
		Type:          OrderType(reo.Type),
		Side:          OrderSide(reo.Side),
		StopPrice:     stopPrice,
		IcebergQty:    icebergQty,
		Time:          t,
	}, nil
}
