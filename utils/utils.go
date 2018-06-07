package utils

import (
	"log"
	"reflect"
	"strconv"
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
