package bitstamp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func sign(str, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(str))

	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}
