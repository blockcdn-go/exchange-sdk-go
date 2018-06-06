package weex

import (
	"crypto/md5"
	"fmt"
	"log"
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
	if from == nil {
		return []sortPair{}
	}
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

func urlEncode(in map[string]interface{}) string {
	if in == nil || len(in) == 0 {
		return ""
	}
	s := sortMap(in)
	return sliceEncode(s)
}

func sliceEncode(s []sortPair) string {
	var str string
	for i := 0; i < len(s); i++ {
		str += s[i].Key + "=" + url.QueryEscape(toString(s[i].Value))
		if i != len(s)-1 {
			str += "&"
		}
	}
	return str
}

func sign(apikey, str string) string {

	h := md5.New()
	h.Write([]byte(str))
	s := fmt.Sprintf("%x", h.Sum(nil))
	return strings.ToUpper(s)
}

func toString(i interface{}) string {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32:
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	}
	return ""
}

func toFloat(i interface{}) float64 {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	var f float64
	switch v.Kind() {
	case reflect.String:
		f, _ = strconv.ParseFloat(v.String(), 64)
		break
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		f = float64(v.Int())
		break
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		f = float64(v.Uint())
		break
	case reflect.Float32, reflect.Float64:
		f = v.Float()
		break
	default:
		log.Printf("toFloat type error %v\n", v.Kind())
		break
	}

	return f
}
