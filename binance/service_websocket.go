package binance

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/global"
	"github.com/gorilla/websocket"
)

func (as *apiService) SubDepth(sreq global.TradeSymbol) (chan global.Depth, error) {
	symbol := strings.ToLower(sreq.Base + sreq.Quote)
	url := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@depth", symbol)
	dial := websocket.DefaultDialer
	if as.proxy != nil {
		dial.Proxy = http.ProxyURL(as.proxy)
	}
	c, _, err := dial.Dial(url, nil)
	if err != nil {
		log.Println("dial:", err)
		return nil, err
	}

	dech := make(chan global.Depth)

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
					// reconnect
					for {
						c, _, err = dial.Dial(url, nil)
						if err == nil {
							log.Println("reconnect success")
							break
						}
						time.Sleep(time.Second * 5)
					}
					continue
				}
				//fmt.Println("binance depth:", string(message))
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
				for _, a := range rawDepth.AskDepthDelta {
					p, err := floatFromString(a[0])
					if err != nil {
						log.Println("wsUnmarshal", err, "body", string(message))
						return
					}
					q, err := floatFromString(a[1])
					if err != nil {
						log.Println("wsUnmarshal", err, "body", string(message))
						return
					}
					de.Asks = append(de.Asks, &Order{
						Price:    p,
						Quantity: q,
					})
				}

				//
				r := global.Depth{
					Base:  sreq.Base,
					Quote: sreq.Quote,
					Asks:  make([]global.DepthPair, 0, 5),
					Bids:  make([]global.DepthPair, 0, 5),
				}
				for _, a := range de.Asks {
					if a.Price == 0. || a.Quantity == 0. {
						continue
					}
					r.Asks = append(r.Asks, global.DepthPair{
						Price: a.Price,
						Size:  a.Quantity,
					})
				}
				for _, b := range de.Bids {
					if b.Price == 0. || b.Quantity == 0. {
						continue
					}
					r.Bids = append(r.Bids, global.DepthPair{
						Price: b.Price,
						Size:  b.Quantity,
					})
				}
				dech <- r
			}
		}
	}()

	return dech, nil
}

func (as *apiService) SubLateTrade(sreq global.TradeSymbol) (chan global.LateTrade, error) {
	symbol := strings.ToLower(sreq.Base + sreq.Quote)
	url := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@aggTrade", symbol)
	dial := websocket.DefaultDialer
	if as.proxy != nil {
		dial.Proxy = http.ProxyURL(as.proxy)
	}
	c, _, err := dial.Dial(url, nil)
	if err != nil {
		log.Println("dial:", err)
		return nil, err
	}

	aggtech := make(chan global.LateTrade)

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
					// reconnect
					for {
						c, _, err = dial.Dial(url, nil)
						if err == nil {
							log.Println("reconnect success")
							break
						}
						time.Sleep(time.Second * 5)
					}
					continue
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
				//////
				ret := global.LateTrade{
					Base:      sreq.Base,
					Quote:     sreq.Quote,
					DateTime:  ae.Timestamp.Format("2006-01-02 03:04:05 PM"),
					Num:       ae.Quantity,
					Price:     ae.Price,
					Total:     ae.Price * ae.Quantity,
					Dircetion: "buy",
				}
				if !ae.BuyerMaker {
					ret.Dircetion = "sell"
				}
				aggtech <- ret
			}
		}
	}()

	return aggtech, nil
}

func (as *apiService) SubTicker(sreq global.TradeSymbol) (chan global.Ticker, error) {
	symbol := strings.ToLower(sreq.Base + sreq.Quote)
	url := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@ticker", symbol)
	dial := websocket.DefaultDialer
	if as.proxy != nil {
		dial.Proxy = http.ProxyURL(as.proxy)
	}
	c, _, err := dial.Dial(url, nil)
	if err != nil {
		log.Println("dial:", err)
		return nil, err
	}

	tk := make(chan global.Ticker)
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
					// reconnect
					for {
						c, _, err = dial.Dial(url, nil)
						if err == nil {
							log.Println("reconnect success")
							break
						}
						time.Sleep(time.Second * 5)
					}
					continue
				}
				rawTicker24 := struct {
					PriceChange        string  `json:"p"`
					PriceChangePercent string  `json:"P"`
					WeightedAvgPrice   string  `json:"w"`
					PrevClosePrice     string  `json:"x"`
					LastPrice          string  `json:"c"`
					BidPrice           string  `json:"b"`
					AskPrice           string  `json:"a"`
					OpenPrice          string  `json:"o"`
					HighPrice          string  `json:"h"`
					LowPrice           string  `json:"l"`
					Volume             string  `json:"v"`
					OpenTime           float64 `json:"O"`
					CloseTime          float64 `json:"C"`
					FirstID            int     `json:"F"`
					LastID             int     `json:"L"`
					Count              int     `json:"n"`
				}{}
				if err := json.Unmarshal(message, &rawTicker24); err != nil {
					log.Println("wsUnmarshal", err, "body", string(message))
					continue
				}

				//fmt.Println("ticker:", string(message))

				pc, err := strconv.ParseFloat(rawTicker24.PriceChange, 64)
				if err != nil {
					continue
				}
				pcPercent, err := strconv.ParseFloat(rawTicker24.PriceChangePercent, 64)
				if err != nil {
					continue
				}
				wap, err := strconv.ParseFloat(rawTicker24.WeightedAvgPrice, 64)
				if err != nil {
					continue
				}
				pcp, err := strconv.ParseFloat(rawTicker24.PrevClosePrice, 64)
				if err != nil {
					continue
				}
				lastPrice, err := strconv.ParseFloat(rawTicker24.LastPrice, 64)
				if err != nil {
					continue
				}
				bp, err := strconv.ParseFloat(rawTicker24.BidPrice, 64)
				if err != nil {
					continue
				}
				ap, err := strconv.ParseFloat(rawTicker24.AskPrice, 64)
				if err != nil {
					continue
				}
				op, err := strconv.ParseFloat(rawTicker24.OpenPrice, 64)
				if err != nil {
					continue
				}
				hp, err := strconv.ParseFloat(rawTicker24.HighPrice, 64)
				if err != nil {
					continue
				}
				lowPrice, err := strconv.ParseFloat(rawTicker24.LowPrice, 64)
				if err != nil {
					continue
				}
				vol, err := strconv.ParseFloat(rawTicker24.Volume, 64)
				if err != nil {
					continue
				}
				ot, err := timeFromUnixTimestampFloat(rawTicker24.OpenTime)
				if err != nil {
					continue
				}
				ct, err := timeFromUnixTimestampFloat(rawTicker24.CloseTime)
				if err != nil {
					continue
				}
				t24 := &Ticker24{
					Symbol:             symbol,
					PriceChange:        pc,
					PriceChangePercent: pcPercent,
					WeightedAvgPrice:   wap,
					PrevClosePrice:     pcp,
					LastPrice:          lastPrice,
					BidPrice:           bp,
					AskPrice:           ap,
					OpenPrice:          op,
					HighPrice:          hp,
					LowPrice:           lowPrice,
					Volume:             vol,
					OpenTime:           ot,
					CloseTime:          ct,
					FirstID:            rawTicker24.FirstID,
					LastID:             rawTicker24.LastID,
					Count:              rawTicker24.Count,
				}

				///
				r := global.Ticker{
					Base:               sreq.Base,
					Quote:              sreq.Quote,
					PriceChange:        t24.PriceChange,
					PriceChangePercent: t24.PriceChangePercent,
					LastPrice:          t24.LastPrice,
					HighPrice:          t24.HighPrice,
					LowPrice:           t24.LowPrice,
					Volume:             t24.Volume,
				}
				tk <- r
			}
		}
	}()

	return tk, nil
}

func (as *apiService) Ticker24Websocket() (chan *Ticker24, error) {
	url := "wss://stream.binance.com:9443/ws/!miniTicker@arr@3000ms"
	dial := websocket.DefaultDialer
	if as.proxy != nil {
		dial.Proxy = http.ProxyURL(as.proxy)
	}
	c, _, err := dial.Dial(url, nil)
	if err != nil {
		log.Println("dial:", err)
		return nil, err
	}

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
					// reconnect
					for {
						c, _, err = dial.Dial(url, nil)
						if err == nil {
							log.Println("reconnect success")
							break
						}
						time.Sleep(time.Second * 5)
					}
					continue
				}
				arrtk := make([]struct {
					LastPrice string  `json:"c"` //
					OpenPrice string  `json:"o"` //
					HighPrice string  `json:"h"` //
					LowPrice  string  `json:"l"` //
					Volume    string  `json:"v"` //
					OpenTime  float64 `json:"E"` //
					Event     string  `json:"e"` //
					Symbol    string  `json:"s"`
				}, 0)
				if err := json.Unmarshal(message, &arrtk); err != nil {
					log.Println("wsUnmarshal", err, "body", string(message))
					continue
				}

				for _, rawTicker24 := range arrtk {

					lastPrice, err := strconv.ParseFloat(rawTicker24.LastPrice, 64)
					if err != nil {
						continue
					}
					op, err := strconv.ParseFloat(rawTicker24.OpenPrice, 64)
					if err != nil {
						continue
					}
					hp, err := strconv.ParseFloat(rawTicker24.HighPrice, 64)
					if err != nil {
						continue
					}
					lowPrice, err := strconv.ParseFloat(rawTicker24.LowPrice, 64)
					if err != nil {
						continue
					}
					vol, err := strconv.ParseFloat(rawTicker24.Volume, 64)
					if err != nil {
						continue
					}
					ot, err := timeFromUnixTimestampFloat(rawTicker24.OpenTime)
					if err != nil {
						continue
					}

					t24 := &Ticker24{
						LastPrice: lastPrice,
						OpenPrice: op,
						HighPrice: hp,
						LowPrice:  lowPrice,
						Volume:    vol,
						OpenTime:  ot,
						Symbol:    rawTicker24.Symbol,
					}
					tk <- t24
				}
			}
		}
	}()

	return tk, nil
}
func (as *apiService) KlineWebsocket(symbol string, intr Interval) (chan *KlineEvent, error) {
	url := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@kline_%s", strings.ToLower(symbol), string(intr))
	dial := websocket.DefaultDialer
	if as.proxy != nil {
		dial.Proxy = http.ProxyURL(as.proxy)
	}
	c, _, err := dial.Dial(url, nil)
	if err != nil {
		log.Println("dial:", err)
		return nil, err
	}

	kech := make(chan *KlineEvent)

	go func() {
		defer c.Close()
		for {
			select {
			case <-as.Ctx.Done():
				log.Println("closing reader")
				return
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Println("wsRead", err)
					// reconnect
					for {
						c, _, err = dial.Dial(url, nil)
						if err == nil {
							log.Println("reconnect success")
							break
						}
						time.Sleep(time.Second * 5)
					}
					continue
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

	return kech, nil
}
func (as *apiService) UserDataWebsocket(listenKey string) (chan *AccountEvent, error) {
	url := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s", listenKey)
	dial := websocket.DefaultDialer
	if as.proxy != nil {
		dial.Proxy = http.ProxyURL(as.proxy)
	}
	c, _, err := dial.Dial(url, nil)
	if err != nil {
		log.Println("dial:", err)
		return nil, err
	}

	aech := make(chan *AccountEvent)

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
				rawAccount := struct {
					Type            string  `json:"e"`
					Time            float64 `json:"E"`
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

	return aech, nil
}
