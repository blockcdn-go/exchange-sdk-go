package coinex

import (
	"errors"
	"strings"

	"github.com/blockcdn-go/exchange-sdk-go/global"
	"github.com/blockcdn-go/exchange-sdk-go/utils"
)

type plainRsp struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func (c *Client) getLateTrade(id interface{}, req global.TradeSymbol) ([]global.LateTrade, error) {
	in := map[string]interface{}{}
	in["market"] = strings.ToUpper(req.Base + req.Quote)
	if id != nil {
		in["last_id"] = id
	}

	data := []map[string]interface{}{}
	r := plainRsp{Data: &data}
	err := c.httpReq("GET", "https://api.coinex.com/v1/market/deals", in, &r, false)
	if err != nil {
		return nil, err
	}
	if r.Code != 0 {
		return nil, errors.New(r.Message)
	}
	ret := []global.LateTrade{}
	for _, l := range data {
		ret = append(ret, global.LateTrade{
			Base:      req.Base,
			Quote:     req.Quote,
			DateTime:  utils.Strftime(l["date"]),
			Num:       utils.ToFloat(l["amount"]),
			Price:     utils.ToFloat(l["price"]),
			Dircetion: utils.ToString(l["type"]),
		})
		ret[len(ret)-1].Total = ret[len(ret)-1].Price * ret[len(ret)-1].Num
		if c.ltid < int64(utils.ToFloat(l["id"])) {
			c.ltid = int64(utils.ToFloat(l["id"]))
		}
	}
	return ret, nil
}
