package gate

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"strconv"

	"github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

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

// MarketList 交易市场详细行情接口
func (c *Client) MarketList() ([]MarketListResponse, error) {
	r := c.newRequest(http.MethodGet, *c.config.RESTHost, "/api2/1/marketlist")
	resp, err := c.doRequest(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，响应码：%d", resp.StatusCode)
	}

	var result struct {
		Result string               `json:"result"`
		Data   []MarketListResponse `json:"data"`
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	extra.RegisterFuzzyDecoders()

	err = jsoniter.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// 获取行情ticker
/*
 	baseVolume: 交易量
    high24hr:24小时最高价
    highestBid:买方最高价
    last:最新成交价
    low24hr:24小时最低价
    lowestAsk:卖方最低价
    percentChange:涨跌百分比
    quoteVolume: 兑换货币交易量
*/

type TickerResponse struct{
	Base		string
	Quote		string
	BaseVolume	float64 `josn:"baseVolume"`
	High24hr	float64 `json:"high24hr"`
	Low24hr		float64 `json:"low24hr"`
	HighestBid	float64 `json:"highestBid"`
	LowestAsk	float64 `json:"lowestAsk"`
	Last		float64 `json:"last"`
	PercentChange float64 `json:"percentChange"`
	QuoteVolume float64 `json:"quoteVolume"`
}

func (c *Client) TickerInfo(base,quote string) (TickerResponse, error) {
	path := fmt.Sprintf("/api2/1/ticker/%s_%s",base,quote)
	var result TickerResponse
	e := c.httpReq("GET",path, "",&result)
	result.Base = base
	result.Quote = quote
	return result, e
}

// 市场深度
/*

*/

type Depth5 struct{
	Base		string
	Quote		string
	AskPirce1	float64
	AskPirce2	float64
	AskPirce3	float64
	AskPirce4	float64
	AskPirce5	float64
	AskSize1	float64
	AskSize2	float64
	AskSize3	float64
	AskSize4	float64
	AskSize5	float64
	//
	BidPrice1	float64
	BidPrice2	float64
	BidPrice3	float64
	BidPrice4	float64
	BidPrice5	float64
	BidSize1	float64
	BidSize2	float64
	BidSize3	float64
	BidSize4	float64
	BidSize5	float64
}

func (c *Client) DepthInfo(base,quote string)(Depth5,error){
	path := fmt.Sprintf("/api2/1/orderBook/%s_%s",base,quote)
	t := struct{
		Asks 	[][]float64	`json:"asks"`
		Bids	[][]float64 `json:"bids"`
	}{}

	e := c.httpReq("GET",path,"",&t)
	if e != nil {
		return Depth5{}, e
	}
	if len(t.Asks) < 5 || len(t.Bids) < 5 {
		return Depth5{}, fmt.Errorf("depth len < 5")
	}
	var r Depth5
	r.Base = base
	r.Quote = quote
	
	r.AskPirce1 = t.Asks[0][0]
	r.AskPirce2 = t.Asks[1][0]
	r.AskPirce3 = t.Asks[2][0]
	r.AskPirce4 = t.Asks[3][0]
	r.AskPirce5 = t.Asks[4][0]
	r.AskSize1 = t.Asks[0][1]
	r.AskSize2 = t.Asks[1][1]
	r.AskSize3 = t.Asks[2][1]
	r.AskSize4 = t.Asks[3][1]
	r.AskSize5 = t.Asks[4][1]
	//
	r.BidPrice1 = t.Bids[0][0]
	r.BidPrice2 = t.Bids[1][0]
	r.BidPrice3 = t.Bids[2][0]
	r.BidPrice4 = t.Bids[3][0]
	r.BidPrice5 = t.Bids[4][0]
	r.BidSize1 = t.Bids[0][1]
	r.BidSize2 = t.Bids[1][1]
	r.BidSize3 = t.Bids[2][1]
	r.BidSize4 = t.Bids[3][1]
	r.BidSize5 = t.Bids[4][1]
	return r, nil
}


//////////////////////////////////////////////////////////////////////////
/// 交易类接口

type Balance struct{
	Available map[string]float64
	Locked    map[string]float64
}

// 获取帐号资金余额
func (c *Client)BalanceInfo()(Balance, error){
	path := "/api2/1/private/balances"
	b := struct{
		Result 		string 		`json:"result"`
		Available 	interface{} `json:"available"`
		Locked 		interface{} `json:"locked"`
	}{}
	e := c.httpReq("POST",path,"",&b)
	if e != nil {
		return Balance{}, e
	}
	if b.Result != "true" {
		return Balance{}, fmt.Errorf("get balances result false")
	}
	var r Balance
	r.Available = make(map[string]float64)
	r.Locked = make(map[string]float64)

	t1, o1 := b.Available.(map[string]string)
	t2, o2 := b.Locked.(map[string]string)
	if o1 {
		fmt.Println("Available cast sucess.")
		if ce := convkv(r.Available, t1); ce != nil {
			return Balance{}, ce
		}
	}
	if o2 {
		fmt.Println("Locked cast success.")
		if ce := convkv(r.Locked, t2); ce != nil {
			return Balance{}, ce
		}
	}
	return r, nil
}


// 获取充值地址
func (c *Client)DepositAddr(currency string)(string, error) {
	path := "/api2/1/private/depositAddress"
	rsp := struct{
		Result 	string 	`json:"result"`
		Addr	string 	`json:"addr"`
		Message string 	`json:"message"`
		Code 	int64 	`json:"code"`
	}{}
	arg := fmt.Sprintf(`currency=%s`,currency)
	e := c.httpReq("POST",path,arg,&rsp)
	if e != nil {
		return "", e
	}
	if rsp.Result != "true" && rsp.Code != 0 {
		return "", fmt.Errorf(rsp.Message)
	}
	return rsp.Addr, nil
}

type DWInfo struct{
	ID 			string `json:"id"`
	Currency 	string `json:"currency"`
	Address		string `json:"address"`
	Amount		string `json:"amount"`
	Txid		string `json:"txid"`
	Timestamp	string `json:"timestamp"`
	Status		string `json:"status"`	//DONE:完成; CANCEL:取消; REQUEST:请求中 
}

// 获取充值提现历史
// return1 充值， return2 提现
func (c *Client)DepositsWithdrawals()([]DWInfo,[]DWInfo,error){
	rsp := struct{
		Result 	string `json:"result"`
		Message string `json:"message"`
		Deposits []DWInfo `json:"deposits"`
		Withdraws []DWInfo `json:"withdraws"`
	}{}
	e := c.httpReq("POST","/api2/1/private/depositsWithdrawals","",&rsp)
	if e != nil {
		return nil, nil, e
	}
	if rsp.Result != "true" {
		return nil, nil, fmt.Errorf(rsp.Message)
	}
	return rsp.Deposits, rsp.Withdraws, nil
}



////////////////////////////////////////////////////////////////////////

func convkv(dst map[string]float64, src map[string]string) error {
	for k,v := range src {
		f, e := strconv.ParseFloat(v, 64)
		if e != nil {
			return e
		}
		dst[k] = f
	}
	return nil
}


//////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////
func (c *Client)httpReq(method, path, data string, v interface{}) error{
	r := c.newRequest(method, *c.config.RESTHost, path)
	if data != "" {
		r.body = strings.NewReader(data)
		defer func() {
			r.body = nil
		}()
	}
	resp, err := c.doRequest(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，响应码：%d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	extra.RegisterFuzzyDecoders()

	err = jsoniter.Unmarshal(body, v)
	if err != nil {
		return err
	}
	return nil
}