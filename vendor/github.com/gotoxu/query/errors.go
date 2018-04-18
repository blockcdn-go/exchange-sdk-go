package query

import (
	"fmt"
	"reflect"
)

// ConversionError 存储了转换错误的详细信息
type ConversionError struct {
	Key   string
	Type  reflect.Type
	Index int
	Err   error
}

func (e ConversionError) Error() string {
	var output string

	if e.Index > 0 {
		output = fmt.Sprintf("url: error converting value for %q", e.Key)
	} else {
		output = fmt.Sprintf("url: error converting value for index %d of %q", e.Index, e.Key)
	}

	if e.Err != nil {
		output = fmt.Sprintf("%s. Details: %s", output, e.Err)
	}

	return output
}

// MultiError 存储了多个编码或解码错误
type MultiError map[string]error

func (e MultiError) Error() string {
	s := ""
	for _, err := range e {
		s = err.Error()
		break
	}

	switch len(e) {
	case 0:
		return "(0 errors)"
	case 1:
		return s
	case 2:
		return s + " (and 1 other error)"
	}

	return fmt.Sprintf("%s (and %d other errors)", s, len(e)-1)
}
