package qbase

import (
	"strings"
)

func split(s string) (string, string) {
	base, quote := s, "error"

	v := strings.Split(s, "/")
	if len(v) >= 2 {
		base, quote = v[0], v[1]
	}

	return strings.ToUpper(base), strings.ToUpper(quote)
}
