package huobi

// TradePair ...
type TradePair struct {
	Base  string `json:"base-currency"`  // 基础币种
	Quote string `json:"quote-currency"` // 计价币种
}

// Account ...
type Account struct {
	AccountID   int64  `json:"id"`
	Status      string `json:"state"` // working：正常, lock：账户被锁定
	AccountType string `json:"type"`  // spot：现货账户
}

// Balance ...
type Balance struct {
	Amount     string `json:"balance"`  // 余额
	CurrencyNo string `json:"currency"` //币种
	BType      string `json:"type"`     // trade: 交易余额，frozen: 冻结余额
}

// InsertOrderReq ...
type InsertOrderReq struct {
	AccountID string `json:"account-id"`      // 账户ID
	Amount    string `json:"amount"`          // 限价表示下单数量, 市价买单时表示买多少钱, 市价卖单时表示卖多少币
	Price     string `json:"price,omitempty"` // 下单价格, 市价单不传该参数
	Source    string `json:"source"`          // 订单来源, api: API调用, margin-api: 借贷资产交易
	Symbol    string `json:"symbol"`          // 交易对, btcusdt, bccbtc......
	OrderType string `json:"type"`            // 订单类型, buy-market: 市价买, sell-market: 市价卖, buy-limit: 限价买, sell-limit: 限价卖
}

// OrderDetail ...
type OrderDetail struct {
	AccountID    int64  `json:"account-id"`        //账户 ID
	Num          int64  `json:"amount"`            //订单数量
	CancelTime   int64  `json:"canceled-at"`       //订单撤销时间
	CreateTime   int64  `json:"created-at"`        //订单创建时间
	TradeNum     string `json:"field-amount"`      //已成交数量
	TradePrice   string `json:"field-cash-amount"` //已成交总金额
	TradeFee     string `json:"field-fees"`        //已成交手续费（买入为币，卖出为钱）
	TradeTime    int64  `json:"finished-at"`       //	最后成交时间
	MatchNo      int64  `json:"id"`                //订单ID
	InsertPrice  string `json:"price"`             //订单价格
	InsertSource string `json:"source"`            //订单来源	api
	OrderStatus  string `json:"state"`             //订单状态	pre-submitted 准备提交, submitting , submitted 已提交, partial-filled 部分成交, partial-canceled 部分成交撤销, filled 完全成交, canceled 已撤销
	Symbol       string `json:"symbol"`            // 交易对	btcusdt, bchbtc, rcneth ...
	OrderType    string `json:"type"`              // 订单类型	buy-market：市价买, sell-market：市价卖, buy-limit：限价买, sell-limit：限价卖
}

// MatchDetail ...
type MatchDetail struct {
	MatchTime   int64  `json:"created-at"`    //成交时间
	MatchNum    string `json:"filled-amount"` //成交数量
	MatchFee    string `json:"filled-fees"`   //成交手续费
	MatchNo     int64  `json:"id"`            //订单成交记录ID
	MatchMarkNo int64  `json:"match-id"`      //撮合ID
	OrderNo     int64  `json:"order-id"`      // 订单 ID
	MatchPrice  string `json:"price"`         //成交价格
	OrderSource string `json:"source"`        //订单来源	api
	Symbol      string `json:"symbol"`        //交易对	btcusdt, bchbtc, rcneth ...
	OrderType   string `json:"type"`          //订单类型	buy-market：市价买, sell-market：市价卖, buy-limit：限价买, sell-limit：限价卖
}
