package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func StringSliceIndexOf(s []string, tar string) int {
	for i, v := range s {
		if v == tar {
			return i
		}
	}
	return -1
}

func Int32SliceIndexOf(s []int32, tar int32) int {
	for i, v := range s {
		if v == tar {
			return i
		}
	}
	return -1
}

func SliceStringToInt32(s []string) []int32 {
	ret := make([]int32, 0)
	for _, v := range s {
		in, err := strconv.Atoi(v)
		if err != nil {
			return nil
		}
		ret = append(ret, int32(in))
	}
	return ret
}
func String2Slice(src string, split ...string) (r []int32) {
	if src == "" {
		return
	}
	var s string
	if len(split) > 0 {
		s = split[0]
	} else {
		s = ","
	}
	arr := strings.Split(src, s)
	for _, v := range arr {
		in, _ := strconv.Atoi(v)
		r = append(r, int32(in))
	}
	return r
}

// 切割2维数组
func String2Slice2(src string, split1, split2 string) (r [][]int32) {
	if src == "" {
		return
	}
	arr := strings.Split(src, split1)
	for _, s := range arr {
		if v := String2Slice(s, split2); len(v) > 0 {
			r = append(r, v)
		}
	}
	return r
}

func ConditionalOperator(op bool, a, b interface{}) interface{} {
	if op {
		return a
	} else {
		return b
	}
}

type ArrInt32 []int32

func (a ArrInt32) Has(i int32) bool {
	for _, v := range a {
		if i == v {
			return true
		}
	}
	return false
}

func (a ArrInt32) String() string {
	var s []string
	for _, v := range a {
		s = append(s, fmt.Sprintf("%d", v))
	}
	return strings.Join(s, ",")
}
