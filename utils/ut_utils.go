package utils

import (
	"log"
	"reflect"
	"strconv"
	"time"
)

// ToString ...
func ToString(i interface{}) string {
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

// ToFloat ...
func ToFloat(i interface{}) float64 {
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

// Ternary 三目运算符
func Ternary(exp bool, a interface{}, b interface{}) interface{} {
	if exp {
		return a
	}
	return b
}

// EndWith 是否以某个字符串结尾
func EndWith(str string, substr string) bool {
	strlen := len(str)
	substrlen := len(substr)
	if substrlen == 0 {
		return true
	}
	if strlen == 0 {
		return false
	}
	s1 := []byte(str)
	s2 := []byte(substr)
	if s1[strlen-1] != s2[substrlen-1] {
		return false
	}
	return EndWith(string(s1[0:strlen-1]), string(s2[0:substrlen-1]))
}

// Strftime 格式化成时间格式
func Strftime(t interface{}) string {
	it := int64(ToFloat(t))
	return time.Unix(it, 0).Format("2006-01-02 03:04:05 PM")
}

// Period2Suffix *m => *min, *h => *hour, *w => *week, *d => *day,
// special60 如果为true 则 1h => 60min
func Period2Suffix(period string, special1h bool) string {
	if special1h && period == "1h" {
		return "60min"
	}
	if EndWith(period, "m") {
		return period + "in"
	}
	if EndWith(period, "h") {
		return period + "our"
	}
	if EndWith(period, "d") {
		return period + "ay"
	}
	if EndWith(period, "w") {
		return period + "eek"
	}
	if EndWith(period, "y") {
		return period + "ear"
	}
	return period
}
