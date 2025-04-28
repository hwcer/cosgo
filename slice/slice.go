package slice

import (
	"fmt"
	"strconv"
	"strings"
)

func Has[T comparable](arr []T, tar T) bool {
	for _, v := range arr {
		if tar == v {
			return true
		}
	}
	return false
}

func IndexOf[T comparable](arr []T, tar T) int {
	for k, v := range arr {
		if v == tar {
			return k
		}
	}
	return -1
}

// String 转换成,分割的字符串
func String[T comparable](arr []T, split ...string) string {
	var s string
	if len(split) > 0 {
		s = split[0]
	} else {
		s = ","
	}
	var str []string
	for _, v := range arr {
		str = append(str, fmt.Sprintf("%v", v))
	}
	return strings.Join(str, s)
}

func Cover(s []string) []int32 {
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
func Split(src string, char ...string) []int32 {
	if src == "" {
		return nil
	}
	if len(char) == 0 {
		char = append(char, ",")
	}
	arr := strings.Split(src, char[0])
	return Cover(arr)
}

func Multiple(src string, char1, char2 string) (r [][]int32) {
	if src == "" {
		return
	}
	arr := strings.Split(src, char1)
	for _, s := range arr {
		if v := Split(s, char2); len(v) > 0 {
			r = append(r, v)
		}
	}
	return r
}
