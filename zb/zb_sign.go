package zb

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
)

func sign(str, key string) string {
	sha := sha1.New()
	sha.Write([]byte(key))
	h := hmac.New(md5.New, []byte(hex.EncodeToString(sha.Sum(nil))))
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
