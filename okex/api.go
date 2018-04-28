package okex

import (
	"errors"
	"fmt"
)

// Login okex登录
// @returns token, error token 用于后续的websocket登录
func (c *Client) Login(username, password string) (string, error) {
	path := fmt.Sprintf("/v3/users/login/login?loginName=%s", username)

	req := struct {
		AreaCode  int32  `json:"areaCode"`
		LoginName string `json:"loginName"`
		Password  string `json:"password"`
	}{}
	req.AreaCode = 86
	req.LoginName = username
	req.Password = password

	rsp := struct {
		Code      int32  `json:"code"`
		DetailMsg string `json:"detailMsg"`
		Msg       string `json:"msg"`
		Data      struct {
			Step2Type int32  `json:"step2Type"`
			Behavior  int32  `json:"behavior"`
			Token     string `json:"token"`
		} `json:"data"`
	}{}
	ex := make(map[string]string)
	// 如果不设置 接口将不会返回数据
	ex["Referer"] = "https://www.okex.cn/account/login"
	e := c.doHTTP("POST", path, req, &rsp, ex)
	if e != nil {
		return "", e
	}
	if rsp.Code != 0 {
		return "", errors.New(rsp.DetailMsg)
	}
	return rsp.Data.Token, nil
}
