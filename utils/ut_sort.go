package utils

import "net/url"

// SortPair ...
type SortPair struct {
	Key   string
	Value interface{}
}

// SortMap ...
func SortMap(from map[string]interface{}) []SortPair {
	if from == nil {
		return []SortPair{}
	}
	cp := make([]SortPair, 0, len(from))
	for k, v := range from {
		cp = append(cp, SortPair{Key: k, Value: v})
	}
	quickSort(cp, 0, len(cp)-1)
	return cp
}

func quickSort(src []SortPair, begin, end int) {
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

// MapEncode ...
func MapEncode(in map[string]interface{}) string {
	if in == nil || len(in) == 0 {
		return ""
	}
	s := SortMap(in)
	return SliceEncode(s)
}

// SliceEncode ...
func SliceEncode(s []SortPair) string {
	var str string
	for i := 0; i < len(s); i++ {
		str += s[i].Key + "=" + url.QueryEscape(ToString(s[i].Value))
		if i != len(s)-1 {
			str += "&"
		}
	}
	return str
}
