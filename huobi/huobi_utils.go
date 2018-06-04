package huobi

import "strings"

// SplitSymbol 分出base 和 quote
func SplitSymbol(symbol string) (string, string) {

	r1 := symbol
	r2 := "error"
	if len(symbol) < 5 {
		return r1, r2
	}

	b := []byte(symbol)

	l3 := string(b[len(b)-3 : len(b)])
	l4 := string(b[len(b)-4 : len(b)])
	if strings.ToUpper(l3) == "BTC" {
		r1 = strings.ToLower(string(b[0 : len(b)-3]))
		r2 = "btc"
	}
	if strings.ToUpper(l3) == "ETH" {
		r1 = strings.ToLower(string(b[0 : len(b)-3]))
		r2 = "eth"
	}
	if strings.ToUpper(l4) == "USDT" {
		r1 = strings.ToLower(string(b[0 : len(b)-4]))
		r2 = "usdt"
	}
	return r1, r2
}
