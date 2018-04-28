package binance

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func (as *apiService) DepthWebsocket(symbol string) (chan *DepthEvent, chan struct{}, error) {
	url := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@depth", strings.ToLower(symbol))
	dial := websocket.DefaultDialer
	if as.proxy != nil {
		dial.Proxy = http.ProxyURL(as.proxy)
	}
	c, _, err := dial.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	done := make(chan struct{})
	dech := make(chan *DepthEvent)

	go func() {
		defer c.Close()
		defer close(done)
		for {
			select {
			case <-as.Ctx.Done():
				log.Println("closing reader ", url)
				return
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Println("wsRead ", err, url)
					return
				}
				rawDepth := struct {
					Type          string          `json:"e"`
					Time          float64         `json:"E"`
					Symbol        string          `json:"s"`
					UpdateID      int             `json:"u"`
					BidDepthDelta [][]interface{} `json:"b"`
					AskDepthDelta [][]interface{} `json:"a"`
				}{}
				if err := json.Unmarshal(message, &rawDepth); err != nil {
					log.Println("wsUnmarshal", err, "body", string(message))
					return
				}
				t, err := timeFromUnixTimestampFloat(rawDepth.Time)
				if err != nil {
					log.Println("wsUnmarshal", err, "body", string(message))
					return
				}
				de := &DepthEvent{
					WSEvent: WSEvent{
						Type:   rawDepth.Type,
						Time:   t,
						Symbol: rawDepth.Symbol,
					},
					UpdateID: rawDepth.UpdateID,
				}
				for _, b := range rawDepth.BidDepthDelta {
					p, err := floatFromString(b[0])
					if err != nil {
						log.Println("wsUnmarshal", err, "body", string(message))
						return
					}
					q, err := floatFromString(b[1])
					if err != nil {
						log.Println("wsUnmarshal", err, "body", string(message))
						return
					}
					de.Bids = append(de.Bids, &Order{
						Price:    p,
						Quantity: q,
					})
				}
				dech <- de
			}
		}
	}()

	go as.exitHandler(c, done)
	return dech, done, nil
}

func (as *apiService) KlineWebsocket(symbol string, intr Interval) (chan *KlineEvent, chan struct{}, error) {
	url := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@kline_%s", strings.ToLower(symbol), string(intr))
	dial := websocket.DefaultDialer
	if as.proxy != nil {
		dial.Proxy = http.ProxyURL(as.proxy)
	}
	c, _, err := dial.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	done := make(chan struct{})
	kech := make(chan *KlineEvent)

	go func() {
		defer c.Close()
		defer close(done)
		for {
			select {
			case <-as.Ctx.Done():
				log.Println("closing reader")
				return
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Println("wsRead", err)
					return
				}
				rawKline := struct {
					Type     string  `json:"e"`
					Time     float64 `json:"E"`
					Symbol   string  `json:"S"`
					OpenTime float64 `json:"t"`
					Kline    struct {
						Interval                 string  `json:"i"`
						FirstTradeID             int64   `json:"f"`
						LastTradeID              int64   `json:"L"`
						Final                    bool    `json:"x"`
						OpenTime                 float64 `json:"t"`
						CloseTime                float64 `json:"T"`
						Open                     string  `json:"o"`
						High                     string  `json:"h"`
						Low                      string  `json:"l"`
						Close                    string  `json:"c"`
						Volume                   string  `json:"v"`
						NumberOfTrades           int     `json:"n"`
						QuoteAssetVolume         string  `json:"q"`
						TakerBuyBaseAssetVolume  string  `json:"V"`
						TakerBuyQuoteAssetVolume string  `json:"Q"`
					} `json:"k"`
				}{}
				if err := json.Unmarshal(message, &rawKline); err != nil {
					log.Println("wsUnmarshal", err, "body", string(message))
					return
				}
				t, err := timeFromUnixTimestampFloat(rawKline.Time)
				if err != nil {
					log.Println("wsUnmarshal", err, "body", rawKline.Time)
					return
				}
				ot, err := timeFromUnixTimestampFloat(rawKline.Kline.OpenTime)
				if err != nil {
					log.Println("wsUnmarshal", err, "body", rawKline.Kline.OpenTime)
					return
				}
				ct, err := timeFromUnixTimestampFloat(rawKline.Kline.CloseTime)
				if err != nil {
					log.Println("wsUnmarshal", err, "body", rawKline.Kline.CloseTime)
					return
				}
				open, err := floatFromString(rawKline.Kline.Open)
				if err != nil {
					log.Println("wsUnmarshal", err, "body", rawKline.Kline.Open)
					return
				}
				cls, err := floatFromString(rawKline.Kline.Close)
				if err != nil {
					log.Println("wsUnmarshal", err, "body", rawKline.Kline.Close)
					return
				}
				high, err := floatFromString(rawKline.Kline.High)
				if err != nil {
					log.Println("wsUnmarshal", err, "body", rawKline.Kline.High)
					return
				}
				low, err := floatFromString(rawKline.Kline.Low)
				if err != nil {
					log.Println("wsUnmarshal", err, "body", rawKline.Kline.Low)
					return
				}
				vol, err := floatFromString(rawKline.Kline.Volume)
				if err != nil {
					log.Println("wsUnmarshal", err, "body", rawKline.Kline.Volume)
					return
				}
				qav, err := floatFromString(rawKline.Kline.QuoteAssetVolume)
				if err != nil {
					log.Println("wsUnmarshal", err, "body", (rawKline.Kline.QuoteAssetVolume))
					return
				}
				tbbav, err := floatFromString(rawKline.Kline.TakerBuyBaseAssetVolume)
				if err != nil {
					log.Println("wsUnmarshal", err, "body", rawKline.Kline.TakerBuyBaseAssetVolume)
					return
				}
				tbqav, err := floatFromString(rawKline.Kline.TakerBuyQuoteAssetVolume)
				if err != nil {
					log.Println("wsUnmarshal", err, "body", rawKline.Kline.TakerBuyQuoteAssetVolume)
					return
				}

				ke := &KlineEvent{
					WSEvent: WSEvent{
						Type:   rawKline.Type,
						Time:   t,
						Symbol: rawKline.Symbol,
					},
					Interval:     Interval(rawKline.Kline.Interval),
					FirstTradeID: rawKline.Kline.FirstTradeID,
					LastTradeID:  rawKline.Kline.LastTradeID,
					Final:        rawKline.Kline.Final,
					Kline: Kline{
						OpenTime:                 ot,
						CloseTime:                ct,
						Open:                     open,
						Close:                    cls,
						High:                     high,
						Low:                      low,
						Volume:                   vol,
						NumberOfTrades:           rawKline.Kline.NumberOfTrades,
						QuoteAssetVolume:         qav,
						TakerBuyBaseAssetVolume:  tbbav,
						TakerBuyQuoteAssetVolume: tbqav,
					},
				}
				kech <- ke
			}
		}
	}()

	go as.exitHandler(c, done)
	return kech, done, nil
}

func (as *apiService) TradeWebsocket(symbol string) (chan *AggTradeEvent, chan struct{}, error) {
	url := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@aggTrade", strings.ToLower(symbol))
	dial := websocket.DefaultDialer
	if as.proxy != nil {
		dial.Proxy = http.ProxyURL(as.proxy)
	}
	c, _, err := dial.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	done := make(chan struct{})
	aggtech := make(chan *AggTradeEvent)

	go func() {
		defer c.Close()
		defer close(done)
		for {
			select {
			case <-as.Ctx.Done():
				log.Println("closing reader ", url)
				return
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Println("wsRead ", err, url)
					return
				}
				rawAggTrade := struct {
					Type         string  `json:"e"`
					Time         float64 `json:"E"`
					Symbol       string  `json:"s"`
					TradeID      int     `json:"a"`
					Price        string  `json:"p"`
					Quantity     string  `json:"q"`
					FirstTradeID int     `json:"f"`
					LastTradeID  int     `json:"l"`
					Timestamp    float64 `json:"T"`
					IsMaker      bool    `json:"m"`
				}{}
				if err := json.Unmarshal(message, &rawAggTrade); err != nil {
					log.Println("wsUnmarshal", err, "body", string(message))
					return
				}
				t, err := timeFromUnixTimestampFloat(rawAggTrade.Time)
				if err != nil {
					log.Println("wsUnmarshal", err, "body", rawAggTrade.Time)
					return
				}

				price, err := floatFromString(rawAggTrade.Price)
				if err != nil {
					log.Println("wsUnmarshal", err, "body", rawAggTrade.Price)
					return
				}
				qty, err := floatFromString(rawAggTrade.Quantity)
				if err != nil {
					log.Println("wsUnmarshal", err, "body", rawAggTrade.Quantity)
					return
				}
				ts, err := timeFromUnixTimestampFloat(rawAggTrade.Timestamp)
				if err != nil {
					log.Println("wsUnmarshal", err, "body", rawAggTrade.Timestamp)
					return
				}

				ae := &AggTradeEvent{
					WSEvent: WSEvent{
						Type:   rawAggTrade.Type,
						Time:   t,
						Symbol: rawAggTrade.Symbol,
					},
					AggTrade: AggTrade{
						ID:           rawAggTrade.TradeID,
						Price:        price,
						Quantity:     qty,
						FirstTradeID: rawAggTrade.FirstTradeID,
						LastTradeID:  rawAggTrade.LastTradeID,
						Timestamp:    ts,
						BuyerMaker:   rawAggTrade.IsMaker,
					},
				}
				aggtech <- ae
			}
		}
	}()

	go as.exitHandler(c, done)
	return aggtech, done, nil
}

func (as *apiService) TickerWebsocket(symbol string) (chan *Ticker24, chan struct{}, error) {
	url := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@ticker", strings.ToLower(symbol))
	dial := websocket.DefaultDialer
	if as.proxy != nil {
		dial.Proxy = http.ProxyURL(as.proxy)
	}
	c, _, err := dial.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	done := make(chan struct{})
	tk := make(chan *Ticker24)
	go func() {
		defer c.Close()
		for {
			select {
			case <-as.Ctx.Done():
				log.Println("closing reader ", url)
				return
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Println("wsRead ", err, url)
					return
				}
				fmt.Println("ticker:", string(message))
			}
		}
	}()

	return tk, done, nil
}

func (as *apiService) UserDataWebsocket(urwr UserDataWebsocketRequest) (chan *AccountEvent, chan struct{}, error) {
	url := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s", urwr.ListenKey)
	dial := websocket.DefaultDialer
	if as.proxy != nil {
		dial.Proxy = http.ProxyURL(as.proxy)
	}
	c, _, err := dial.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	done := make(chan struct{})
	aech := make(chan *AccountEvent)

	go func() {
		defer c.Close()
		defer close(done)
		for {
			select {
			case <-as.Ctx.Done():
				log.Println("closing reader ", url)
				return
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Println("wsRead ", err, url)
					return
				}
				rawAccount := struct {
					Type            string  `json:"e"`
					Time            float64 `json:"E"`
					OpenTime        float64 `json:"t"`
					MakerCommision  int64   `json:"m"`
					TakerCommision  int64   `json:"t"`
					BuyerCommision  int64   `json:"b"`
					SellerCommision int64   `json:"s"`
					CanTrade        bool    `json:"T"`
					CanWithdraw     bool    `json:"W"`
					CanDeposit      bool    `json:"D"`
					Balances        []struct {
						Asset            string `json:"a"`
						AvailableBalance string `json:"f"`
						Locked           string `json:"l"`
					} `json:"B"`
				}{}
				if err := json.Unmarshal(message, &rawAccount); err != nil {
					log.Println("wsUnmarshal", err, "body", string(message))
					return
				}
				t, err := timeFromUnixTimestampFloat(rawAccount.Time)
				if err != nil {
					log.Println("wsUnmarshal", err, "body", rawAccount.Time)
					return
				}

				ae := &AccountEvent{
					WSEvent: WSEvent{
						Type: rawAccount.Type,
						Time: t,
					},
					Account: Account{
						MakerCommision:  rawAccount.MakerCommision,
						TakerCommision:  rawAccount.TakerCommision,
						BuyerCommision:  rawAccount.BuyerCommision,
						SellerCommision: rawAccount.SellerCommision,
						CanTrade:        rawAccount.CanTrade,
						CanWithdraw:     rawAccount.CanWithdraw,
						CanDeposit:      rawAccount.CanDeposit,
					},
				}
				for _, b := range rawAccount.Balances {
					free, err := floatFromString(b.AvailableBalance)
					if err != nil {
						log.Println("wsUnmarshal", err, "body", b.AvailableBalance)
						return
					}
					locked, err := floatFromString(b.Locked)
					if err != nil {
						log.Println("wsUnmarshal", err, "body", b.Locked)
						return
					}
					ae.Balances = append(ae.Balances, &Balance{
						Asset:  b.Asset,
						Free:   free,
						Locked: locked,
					})
				}
				aech <- ae
			}
		}
	}()

	go as.exitHandler(c, done)
	return aech, done, nil
}

func (as *apiService) exitHandler(c *websocket.Conn, done chan struct{}) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	defer c.Close()

	for {
		select {
		//case t := <-ticker.C:
		// err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
		// if err != nil {
		// 	log.Println("wsWrite", err)
		// 	return
		// }
		case <-as.Ctx.Done():
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			log.Println("closing connection")
			return
		}
	}
}
