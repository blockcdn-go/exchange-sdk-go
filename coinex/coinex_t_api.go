package coinex

import (
	"errors"

	"github.com/blockcdn-go/exchange-sdk-go/utils"

	"github.com/blockcdn-go/exchange-sdk-go/global"
)

// GetFund 获取帐号资金余额
func (c *Client) GetFund(global.FundReq) ([]global.Fund, error) {
	in := map[string]interface{}{}
	data := map[string]map[string]interface{}{}
	r := plainRsp{Data: &data}
	err := c.httpReq("GET", "https://api.coinex.com/v1/balance/", in, &r, true)
	if err != nil {
		return nil, err
	}
	if r.Code != 0 {
		return nil, errors.New(r.Message)
	}
	ret := []global.Fund{}
	for c, v := range data {
		cc := global.Fund{
			Base:      c,
			Available: utils.ToFloat(v["available"]),
			Frozen:    utils.ToFloat(v["frozen"]),
		}
		ret = append(ret, cc)
	}
	return ret, nil
}
