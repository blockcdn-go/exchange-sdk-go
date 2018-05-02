package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Service represents service layer for Binance API.
//
// The main purpose for this layer is to be replaced with dummy implementation
// if necessary without need to replace Binance instance.
type Service interface {
	// Ping tests connectivity.
	Ping() error
	// Time returns server time.
	Time() (time.Time, error)
	// OrderBook returns list of orders.
	OrderBook(obr OrderBookRequest) (*OrderBook, error)
	// AggTrades returns compressed/aggregate list of trades.
	AggTrades(atr AggTradesRequest) ([]*AggTrade, error)
	// Klines returns klines/candlestick data.
	Klines(kr KlinesRequest) ([]*Kline, error)
	// Ticker24 returns 24hr price change statistics.
	Ticker24(symbol string) (*Ticker24, error)
	// TickerAllPrices returns ticker data for symbols.
	TickerAllPrices() ([]*PriceTicker, error)
	// TickerAllBooks returns tickers for all books.
	TickerAllBooks() ([]*BookTicker, error)

	// NewOrder places new order and returns ProcessedOrder.
	NewOrder(nor NewOrderRequest) (*ProcessedOrder, error)
	// NewOrder places testing order.
	NewOrderTest(nor NewOrderRequest) error
	// QueryOrder returns data about existing order.
	QueryOrder(qor QueryOrderRequest) (*ExecutedOrder, error)
	// CancelOrder cancels order.
	CancelOrder(cor CancelOrderRequest) (*CanceledOrder, error)
	// OpenOrders returns list of open orders.
	OpenOrders(oor OpenOrdersRequest) ([]*ExecutedOrder, error)
	// AllOrders returns list of all previous orders.
	AllOrders(aor AllOrdersRequest) ([]*ExecutedOrder, error)

	// Account returns account data.
	Account(ar AccountRequest) (*Account, error)
	// MyTrades list user's trades.
	MyTrades(mtr MyTradesRequest) ([]*Trade, error)
	// Withdraw executes withdrawal.
	Withdraw(wr WithdrawRequest) (*WithdrawResult, error)
	// DepositHistory lists deposit data.
	DepositHistory(hr HistoryRequest) ([]*Deposit, error)
	// WithdrawHistory lists withdraw data.
	WithdrawHistory(hr HistoryRequest) ([]*Withdrawal, error)

	// StartUserDataStream starts stream and returns Stream with ListenKey.
	StartUserDataStream() (string, error)
	// KeepAliveUserDataStream prolongs stream livespan.
	KeepAliveUserDataStream(listenKey string) error
	// CloseUserDataStream closes opened stream.
	CloseUserDataStream(listenKey string) error

	DepthWebsocket(symbol string) (chan *DepthEvent, error)
	KlineWebsocket(symbol string, intr Interval) (chan *KlineEvent, error)
	TradeWebsocket(symbol string) (chan *AggTradeEvent, error)
	TickerWebsocket(symbol string) (chan *Ticker24, error)
	Ticker24Websocket() (chan *Ticker24, error)
	UserDataWebsocket(listenKey string) (chan *AccountEvent, error)
}

type apiService struct {
	URL    string
	APIKey string
	APISec string
	Signer Signer
	Ctx    context.Context
	proxy  *url.URL
}

// NewAPIService creates instance of Service.
//
// If logger or ctx are not provided, NopLogger and Background context are used as default.
// You can use context for one-time request cancel (e.g. when shutting down the app).
func NewAPIService(ctx context.Context, url, apiKey, apiSec string, pxy *url.URL) Service {
	if ctx == nil {
		ctx = context.Background()
	}

	return &apiService{
		URL:    url,
		APIKey: apiKey,
		APISec: apiSec,
		proxy:  pxy,
		Signer: &HmacSigner{
			Key: []byte(apiSec),
		},
		Ctx: ctx,
	}
}

func (as *apiService) request(method string, path string, params map[string]string,
	rsp interface{}, apiKey bool, sign bool) error {
	transport := &http.Transport{}
	if as.proxy != nil {
		transport.Proxy = http.ProxyURL(as.proxy)
	}
	client := &http.Client{
		Transport: transport,
	}

	url := fmt.Sprintf("%s/%s", as.URL, path)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}
	req.WithContext(as.Ctx)

	q := req.URL.Query()
	for key, val := range params {
		q.Add(key, val)
	}
	if apiKey {
		req.Header.Add("X-MBX-APIKEY", as.APIKey)
	}
	if sign {
		log.Println("queryString", q.Encode())
		q.Add("signature", as.Signer.Sign([]byte(q.Encode())))
		log.Println("signature", as.Signer.Sign([]byte(q.Encode())))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	textRes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return warpError(err, "unable to read response from allOrders.get")
	}
	fmt.Println("binance http msg:", string(textRes))
	if resp.StatusCode != 200 {
		return as.handleError(textRes)
	}
	if err := json.Unmarshal(textRes, rsp); err != nil {
		return warpError(err, "allOrders unmarshal failed")
	}

	return nil
}
