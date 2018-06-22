package coinex

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

func sign(str, key string) string {
	h := md5.New()
	h.Write([]byte(str))

	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}
