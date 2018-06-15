package coinegg

import "github.com/blockcdn-go/exchange-sdk-go/global"

// GetFund 获取帐号资金余额
func (c *Client) GetFund(global.FundReq) ([]global.Fund, error) {
	arg := map[string]interface{}{}
	r := map[string]interface{}{}
	err := c.httpReq("POST", "https://api.coinegg.com/api/v1/balance/", arg, &r, true)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
