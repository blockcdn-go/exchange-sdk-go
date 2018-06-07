package huobi

import (
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strings"
)

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
		r1 = strings.ToUpper(string(b[0 : len(b)-3]))
		r2 = "BTC"
	}
	if strings.ToUpper(l3) == "ETH" {
		r1 = strings.ToUpper(string(b[0 : len(b)-3]))
		r2 = "ETH"
	}
	if strings.ToUpper(l4) == "USDT" {
		r1 = strings.ToUpper(string(b[0 : len(b)-4]))
		r2 = "USDT"
	}
	return r1, r2
}

type sortPair struct {
	Key   string
	Value string
}

type sortPairSlice []sortPair

func (s sortPairSlice) Len() int {
	return len(s)
}
func (s sortPairSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s sortPairSlice) Less(i, j int) bool {
	return s[i].Key < s[j].Key
}

func mapSort(m map[string]string) sortPairSlice {
	var r sortPairSlice
	for k, v := range m {
		r = append(r, sortPair{k, v})
	}
	sort.Sort(r)
	return r
}

func valURIQuery(s sortPairSlice) string {
	var str string
	for i := 0; i < len(s); i++ {
		str += s[i].Key + "=" + url.QueryEscape(s[i].Value)
		if i != len(s)-1 {
			str += "&"
		}
	}
	return str
}

func if2map(i interface{}) map[string]string {
	r := make(map[string]string)

	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		panic("if2map Only support for structural transformation")
	}
	t := v.Type()

	for idx := 0; idx < v.NumField(); idx++ {

		name, ops := parseTag(t.Field(idx))
		if name == "-" {
			continue
		}
		switch v.Field(idx).Type().Kind() {
		case reflect.String:
			val := v.Field(idx).String()
			eqz := v.Field(idx).Interface() == reflect.Zero(reflect.TypeOf(v.Field(idx).Interface())).Interface()
			if contanis(ops, "omitempty") && eqz {
				continue
			}
			r[name] = val
		// The current parameters are only string type
		default:
			panic(fmt.Sprint("if2map Do not support ", v.Field(idx).Type()))
		}
	}
	return r
}

func parseTag(field reflect.StructField) (string, []string) {

	var ops []string
	name := field.Name

	if tag := field.Tag.Get("json"); tag != "" {
		vstr := strings.Split(tag, ",")
		if len(vstr) > 0 {
			name, ops = vstr[0], vstr[1:]
		}
	}
	return name, ops
}

func contanis(src []string, sub string) bool {
	for _, v := range src {
		if v == sub {
			return true
		}
	}
	return false
}
