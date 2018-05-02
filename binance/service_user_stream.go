package binance

func (as *apiService) StartUserDataStream() (string, error) {
	params := make(map[string]string)
	rsp := struct {
		ListenKey string `json:"listenKey"`
	}{}
	err := as.request("POST", "api/v1/userDataStream", params, &rsp, true, false)
	if err != nil {
		return "", err
	}
	return rsp.ListenKey, nil
}
func (as *apiService) KeepAliveUserDataStream(listenKey string) error {
	params := make(map[string]string)
	params["listenKey"] = listenKey

	err := as.request("PUT", "api/v1/userDataStream", params, nil, true, false)
	if err != nil {
		return err
	}
	return nil
}
func (as *apiService) CloseUserDataStream(listenKey string) error {
	params := make(map[string]string)
	params["listenKey"] = listenKey

	err := as.request("DELETE", "api/v1/userDataStream", params, nil, true, false)
	if err != nil {
		return err
	}
	return nil
}
