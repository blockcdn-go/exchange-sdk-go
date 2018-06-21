package coinex

import (
	"strings"
)

func split(s string) (string, string) {
	return recursionSplit([]byte(strings.ToUpper(s)), 0)
}

func recursionSplit(str []byte, off int) (string, string) {
	if len(str) == off {
		return string(str), "error"
	}
	prefix := string(str[0:off])
	suffix := string(str[off:])
	if suffix == "BTC" || suffix == "BCH" || suffix == "USDT" {
		return prefix, suffix
	}
	return recursionSplit(str, off+1)
}
