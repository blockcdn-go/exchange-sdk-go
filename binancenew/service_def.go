package binance

import (
	"context"
	"fmt"
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
	Ping() error
	Time() (time.Time, error)
	OrderBook(obr OrderBookRequest) (*OrderBook, error)
	AggTrades(atr AggTradesRequest) ([]*AggTrade, error)
	Klines(kr KlinesRequest) ([]*Kline, error)
	Ticker24(tr TickerRequest) (*Ticker24, error)
	TickerAllPrices() ([]*PriceTicker, error)
	TickerAllBooks() ([]*BookTicker, error)

	NewOrder(or NewOrderRequest) (*ProcessedOrder, error)
	NewOrderTest(or NewOrderRequest) error
	QueryOrder(qor QueryOrderRequest) (*ExecutedOrder, error)
	CancelOrder(cor CancelOrderRequest) (*CanceledOrder, error)
	OpenOrders(oor OpenOrdersRequest) ([]*ExecutedOrder, error)
	AllOrders(aor AllOrdersRequest) ([]*ExecutedOrder, error)

	Account(ar AccountRequest) (*Account, error)
	MyTrades(mtr MyTradesRequest) ([]*Trade, error)
	Withdraw(wr WithdrawRequest) (*WithdrawResult, error)
	DepositHistory(hr HistoryRequest) ([]*Deposit, error)
	WithdrawHistory(hr HistoryRequest) ([]*Withdrawal, error)

	StartUserDataStream() (*Stream, error)
	KeepAliveUserDataStream(s *Stream) error
	CloseUserDataStream(s *Stream) error

	DepthWebsocket(dwr DepthWebsocketRequest) (chan *DepthEvent, chan struct{}, error)
	KlineWebsocket(kwr KlineWebsocketRequest) (chan *KlineEvent, chan struct{}, error)
	TradeWebsocket(twr TradeWebsocketRequest) (chan *AggTradeEvent, chan struct{}, error)
	UserDataWebsocket(udwr UserDataWebsocketRequest) (chan *AccountEvent, chan struct{}, error)
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

func (as *apiService) request(method string, endpoint string, params map[string]string,
	apiKey bool, sign bool) (*http.Response, error) {
	transport := &http.Transport{}
	if as.proxy != nil {
		transport.Proxy = http.ProxyURL(as.proxy)
	}
	client := &http.Client{
		Transport: transport,
	}

	url := fmt.Sprintf("%s/%s", as.URL, endpoint)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return resp, nil
}
