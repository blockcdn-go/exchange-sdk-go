package bitstamp

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
)

func sign(str, key string) string {
	m := md5.New()
	m.Write([]byte(key))
	mkey := hex.EncodeToString(m.Sum(nil))

	hs := hmac.New(sha256.New, []byte(mkey))
	hs.Write([]byte(str))
	return hex.EncodeToString(hs.Sum(nil))
}
