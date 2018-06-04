package weex

import (
	"crypto/hmac"
	"crypto/md5"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type sortPair struct {
	Key   string
	Value interface{}
}

func sortMap(from map[string]interface{}) []sortPair {
	cp := make([]sortPair, 0, len(from))
	for k, v := range from {
		cp = append(cp, sortPair{Key: k, Value: v})
	}
	quickSort(cp, 0, len(cp)-1)
	return cp
}

func quickSort(src []sortPair, begin, end int) {
	if begin >= end {
		return
	}
	i := begin
	j := end
	x := src[begin]

	for i < j {
		//从右到左找到第一个小于x的数
		for i < j && src[j].Key >= x.Key {
			j--
		}
		if i < j {
			src[i] = src[j]
			i++
		}
		//从左往右找到第一个大于x的数
		for i < j && src[i].Key <= x.Key {
			i++
		}
		if i < j {
			src[j] = src[i]
			j--
		}
	}
	//i = j的时候，将x填入中间位置
	src[i] = x
	quickSort(src, begin, i-1)
	quickSort(src, i+1, end)
}

func sign(apikey, apisec string, i map[string]interface{}) (string, string) {
	afterSort := sortMap(i)

	var str string
	for i := 0; i < len(afterSort); i++ {
		str += afterSort[i].Key + "=" + url.QueryEscape(toString(afterSort[i].Value))
		if i != len(afterSort)-1 {
			str += "&"
		}
	}
	h := hmac.New(md5.New, []byte(apikey))
	h.Write([]byte(str))
	s := string(h.Sum(nil))
	return strings.ToUpper(s), str
}

func toString(i interface{}) string {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32:
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	}
	return ""
}
