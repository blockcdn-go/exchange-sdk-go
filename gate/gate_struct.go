/*
gate 返回数据的结构体
有可能存在一些字段和gate文档上写的不一样，如果以后需要其他的数据需要对比返回的json字符串
*/

package gate

// MarketListResponse 是MarketList接口的返回值
type MarketListResponse struct {
	No          int     `json:"no"`
	Symbol      string  `json:"symbol"`
	Name        string  `json:"name"`
	NameEn      string  `json:"name_en"`
	NameCn      string  `json:"name_cn"`
	Pair        string  `json:"pair"`
	Rate        string  `json:"rate"`
	VolA        float64 `json:"vol_a"`
	VolB        string  `json:"vol_b"`
	CurrA       string  `json:"curr_a"`
	CurrB       string  `json:"curr_b"`
	CurrSuffix  string  `json:"curr_suffix"`
	RatePercent string  `json:"rate_percent"`
	Trend       string  `json:"trend"`
	Supply      int64   `json:"supply"`
	MarketCap   string  `json:"marketcap"`
}

// TickerResponse ...
type TickerResponse struct {
	Base          string  `json:"base"`
	Quote         string  `json:"quote"`
	BaseVolume    float64 `josn:"baseVolume"`    // 交易量
	High24hr      float64 `json:"high24hr"`      // 24小时最高价
	Low24hr       float64 `json:"low24hr"`       // 24小时最低价
	HighestBid    float64 `json:"highestBid"`    // 买方最高价
	LowestAsk     float64 `json:"lowestAsk"`     // 卖方最低价
	Last          float64 `json:"last"`          // 最新成交价
	PercentChange float64 `json:"percentChange"` // 涨跌百分比
	QuoteVolume   float64 `json:"quoteVolume"`   // 兑换货币交易量
}

// PSpair 深度行情的价格和手数对
type PSpair struct {
	Price float64 `json:"price"`
	Size  float64 `json:"size"`
}

// Depth5 ...
type Depth5 struct {
	Base  string   `json:"base"`
	Quote string   `json:"quote"`
	Asks  []PSpair `json:"asks"` // 卖方
	Bids  []PSpair `json:"bids"` // 买方
}

// Balance ...
type Balance struct {
	Available map[string]float64
	Locked    map[string]float64
}

// DWInfo ...
type DWInfo struct {
	ID        string `json:"id"`
	Currency  string `json:"currency"`
	Address   string `json:"address"`
	Amount    string `json:"amount"`
	Txid      string `json:"txid"`
	Timestamp string `json:"timestamp"`
	Status    string `json:"status"` //DONE:完成; CANCEL:取消; REQUEST:请求中
}

// InsertOrderRsp ...
type InsertOrderRsp struct {
	Result      string  `json:"result"`      // "true" 表示调用成功
	OrderNo     string  `json:"orderNumber"` // orderNumber可用于查询，取消订单。
	InsertPrice float64 `json:"rate"`
	Direction   int     `json:"_"`
	LeftNum     float64 `json:"leftAmount"`
	FilledNum   float64 `json:"filledAmount"`
	FilledPrice float64 `json:"filledRate"`
	Msg         string  `json:"message"`
}

// OrderInfo ...
type OrderInfo struct {
	OrderNo     string  `json:"orderNumber"`   //
	Status      string  `json:"status"`        // 订单状态 cancelled已取消 done已完成
	Symbol      string  `json:"currencyPair"`  //  交易对
	Dircetion   string  `json:"type"`          //买卖类型 sell卖出, buy买入
	TradePrice  float64 `json:"rate"`          //价格
	TradeNum    string  `json:"amount"`        //买卖数量
	InsertPrice float64 `json:"initialRate"`   //下单价格
	InsertNum   string  `json:"initialAmount"` //下单量
}

// HangingOrder ...
type HangingOrder struct {
	LeftNum     string `json:"amount"`        // 订单总数量 剩余未成交数量
	Symbol      string `json:"currencyPair"`  // 订单交易对
	TradeNum    string `json:"filledAmount"`  // 已成交量
	TradePrice  string `json:"filledRate"`    // 成交价格
	InsertNum   string `json:"initialAmount"` // 下单量
	InsertPrice string `json:"initialRate"`   // 下单价格
	OrderNo     string `json:"orderNumber"`   // 订单号
	Price       string `json:"rate"`          // 交易单价
	Status      string `json:"status"`        // 订单状态
	Timestamp   string `json:"timestamp"`     // 时间戳
	Total       string `json:"total"`         //总计
	Direction   string `json:"type"`          // 买卖类型 buy:买入;sell:卖出
}

// Match ...
type Match struct {
	OrderNo    string `json:"orderid"` // 订单id
	Symbol     string `json:"pair"`    // 交易对
	Direction  string `json:"type"`    // 买卖类型
	TradePrice string `json:"rate"`    // 买卖价格
	TradeNum   string `json:"amount"`  // 订单买卖币种数量
	//TradeTime string `json:"time"`	// 订单时间
	TradeTime string `json:"time_unix"` // 订单unix时间戳
}
