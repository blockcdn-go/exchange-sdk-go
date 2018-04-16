package gate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

	fmt.Println(string(body))

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}
