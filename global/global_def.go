package global

// TradeSymbol 交易对
type TradeSymbol struct {
	Base  string `json:"base"`
	Quote string `json:"quote"`
}

// Ticker 实时行情数据
type Ticker struct {
	Base               string  `json:"base"`                 // eg BTC
	Quote              string  `json:"quote"`                // eg USDT
	PriceChange        float64 `json:"price_change"`         // 涨跌值
	PriceChangePercent float64 `json:"price_change_percent"` // 涨跌幅
	LastPrice          float64 `json:"lastprice"`            // 最新价
	HighPrice          float64 `json:"highprice"`            // 最高价
	LowPrice           float64 `json:"lowprice"`             // 最低价
	Volume             float64 `json:"volume"`               // 成交量
}

// DepthPair ...
type DepthPair struct {
	Price float64 `json:"price"` // 价格
	Size  float64 `json:"size"`  // 手数
}

// Depth 深度行情数据
type Depth struct {
	Base  string      `json:"base"`  // eg BTC
	Quote string      `json:"quote"` // eg USDT
	Asks  []DepthPair `json:"asks"`  // 卖
	Bids  []DepthPair `json:"bids"`  // 买
}

// LateTrade 最近成交记录
type LateTrade struct {
	Base      string  `json:"base"`   // eg BTC
	Quote     string  `json:"quote"`  // eg USDT
	DateTime  string  `json:"date"`   // 订单时间
	Num       float64 `json:"amount"` // 成交币种数量
	Price     float64 `json:"rate"`   // 币种单价
	Dircetion string  `json:"type"`   // 买卖类型, buy买 sell卖
	Total     float64 `json:"total"`  // 订单总额
}

// KlineReq 请求查询k线数据
type KlineReq struct {
	Base   string `json:"base"`
	Quote  string `json:"quote"`
	Period string `json:"period"`
	Count  int64  `json:"count"`
	Begin  string `json:"begin"`
	End    string `json:"end"`
}

// Kline K线数据
type Kline struct {
	Base      string  `json:"base"`   // eg BTC
	Quote     string  `json:"quote"`  // eg USDT
	Timestamp int64   `json:"time"`   // 时间戳
	Open      float64 `json:"open"`   // 最高价
	Low       float64 `json:"low"`    // 最低价
	High      float64 `json:"high"`   // 开盘价
	Close     float64 `json:"close"`  // 收盘价
	Volume    float64 `json:"volume"` // 成交量
}

// FundReq 请求查询资金账户
type FundReq struct {
	AccountID string `json:"accountid"`
}

// Fund 资金账户情况
type Fund struct {
	Base      string  `json:"base"`      // e.g BTC
	Available float64 `json:"available"` // 可用
	Frozen    float64 `json:"frozen"`    // 冻结
}

/////////////////////////////////////////////////////////////////

// InsertReq 请求下单参数
type InsertReq struct {
	APIKey    string  `json:"apikey"` // weex 需要
	Base      string  `json:"base"`
	Quote     string  `json:"quote"`
	Price     float64 `json:"price"`
	Num       float64 `json:"num"`
	Type      int     `json:"type"`      // 0 - limit, 1- market
	Direction int     `json:"direction"` // 0 - buy, 1- sell
}

// InsertRsp 请求下单返回
type InsertRsp struct {
	OrderNo string `json:"orderno"`
}

// StatusReq 查询订单状态
type StatusReq struct {
	APIKey  string `json:"apikey"` // weex 需要
	Base    string `json:"base"`
	Quote   string `json:"quote"`
	OrderNo string `json:"orderno"`
}

const (

	// FAILED 下单失败
	FAILED = (100)
	// HANGING 挂起 包含未成交和部分成交
	HANGING = (200)
	// HALFTRADE 部分成交
	HALFTRADE = (300)
	// COMPLETETRADE 已成交 订单完全成交
	COMPLETETRADE = (301)
	// CANCELING 取消中
	CANCELING = (400)
	// CANCELED 已取消
	CANCELED = (401)
)

// StatusRsp 订单状态
type StatusRsp struct {
	TradePrice float64 `json:"tradeprice"`
	TradeNum   float64 `json:"tradenum"`
	Status     int     `json:"status"`
	StatusMsg  string  `json:"statusmsg"`
}

// CancelReq 撤单请求参数
type CancelReq struct {
	APIKey  string `json:"apikey"` // weex 需要
	Base    string `json:"base"`
	Quote   string `json:"quote"`
	OrderNo string `json:"orderno"`
}

// WSif websocket实时推送需要实现的接口
type WSif interface {
	// 订阅ticker
	SubTicker(TradeSymbol) (chan Ticker, error)
	// 订阅深度行情
	SubDepth(TradeSymbol) (chan Depth, error)
	// 订阅最近成交
	SubLateTrade(TradeSymbol) (chan LateTrade, error)
}

// APIif 各个交易所需要实现的接口
type APIif interface {
	//////////////////////////////////////////////////////////////
	// 查询所有交易对
	GetAllSymbol() ([]TradeSymbol, error)
	// 查询kline数据
	GetKline(KlineReq) ([]Kline, error)

	//////////////////////////////////////////////////////////////
	// 获取资金信息
	GetFund(FundReq) ([]Fund, error)
	// 下单,只支持限价和市价
	InsertOrder(InsertReq) (InsertRsp, error)
	// 获取订单状态
	OrderStatus(StatusReq) (StatusRsp, error)
	// 取消订单
	CancelOrder(CancelReq) error
}
