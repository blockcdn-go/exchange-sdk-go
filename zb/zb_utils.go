package zb

import (
	"strings"

	"github.com/blockcdn-go/exchange-sdk-go/utils"
)

func split(s string) (string, string) {
	base, quote := s, "error"
	v := strings.Split(s, "/")
	if len(v) >= 2 {
		base, quote = v[0], v[1]
	}
	return strings.ToUpper(base), strings.ToUpper(quote)
}

func split2(s string) (string, string) {
	base, quote := s, "error"

	v := strings.Split(s, "_")
	if len(v) >= 2 {
		if utils.EndWith(strings.ToUpper(v[0]), "QC") {
			base = string([]byte(v[0])[0 : len(v[0])-2])
			quote = "QC"
		} else if utils.EndWith(strings.ToUpper(v[0]), "ZB") {
			base = string([]byte(v[0])[0 : len(v[0])-2])
			quote = "ZB"
		} else if utils.EndWith(strings.ToUpper(v[0]), "USDT") {
			base = string([]byte(v[0])[0 : len(v[0])-4])
			quote = "USDT"
		} else if utils.EndWith(strings.ToUpper(v[0]), "BTC") {
			base = string([]byte(v[0])[0 : len(v[0])-2])
			quote = "BTC"
		}
	}

	return strings.ToUpper(base), strings.ToUpper(quote)
}

func split3(s string) (string, string) {
	base, quote := s, "error"
	v := strings.Split(s, "_")
	if len(v) >= 2 {
		base, quote = v[0], v[1]
	}
	return strings.ToUpper(base), strings.ToUpper(quote)
}
