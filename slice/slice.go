package slice

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"math/rand"
	"strconv"
	"strings"
)

func Min[T constraints.Ordered](nums []T) (r T) {
	if len(nums) == 0 {
		return
	}
	for i, num := range nums {
		if i == 0 || num < r {
			r = num
		}
	}
	return
}

func Max[T constraints.Ordered](nums []T) (r T) {
	if len(nums) == 0 {
		return
	}
	for i, num := range nums {
		if i == 0 || num > r {
			r = num
		}
	}
	return
}

func Has[T comparable](arr []T, tar T) bool {
	for _, v := range arr {
		if tar == v {
			return true
		}
	}
	return false
}

func Roll[T constraints.Ordered](nums []T) (r T) {
	l := len(nums)
	if l == 0 {
		return
	}
	i := rand.Int31n(int32(l) - 1)
	return nums[i]
}

func IndexOf[T comparable](arr []T, tar T) int {
	for k, v := range arr {
		if v == tar {
			return k
		}
	}
	return -1
}

// Unrepeated 去重
func Unrepeated[T comparable](arr []T) []T {
	r := make([]T, 0)
	s := make(map[T]struct{})
	for _, v := range arr {
		if _, ok := s[v]; !ok {
			r = append(r, v)
			s[v] = struct{}{}
		}
	}
	return r
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

func ParseInt32(v string) (int32, error) {
	if v == "" {
		return 0, nil
	}
	in, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}
	return int32(in), nil
}
func ParseInt64(v string) (int64, error) {
	if v == "" {
		return 0, nil
	}
	in, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}
	return int64(in), nil
}

// Cover 字符串切片转int32切片
// clean 去重
func Cover(s []string, unrepeated ...bool) []int32 {
	ret := make([]int32, 0)
	flag := false
	if len(unrepeated) > 0 {
		flag = unrepeated[0]
	}
	exist := make(map[string]struct{})

	for _, v := range s {
		if _, ok := exist[v]; ok && flag {
			continue
		} else if flag {
			exist[v] = struct{}{}
		}
		in, err := ParseInt32(v)
		if err != nil {
			return nil
		}
		ret = append(ret, in)
	}
	return ret
}

func Split(src string, char ...string) []string {
	if src == "" {
		return nil
	}
	if len(char) == 0 {
		char = append(char, ",")
	}
	return strings.Split(src, char[0])
}

func SplitInt32(src string, char ...string) (r []int32) {
	arr := Split(src, char...)
	for _, v := range arr {
		in, _ := ParseInt32(v)
		r = append(r, in)
	}
	return
}
func SplitInt64(src string, char ...string) (r []int64) {
	arr := Split(src, char...)
	for _, v := range arr {
		in, _ := ParseInt64(v)
		r = append(r, in)
	}
	return
}

// SplitAndUnrepeated 切割并去重
func SplitAndUnrepeated(src string, char ...string) []int32 {
	if src == "" {
		return nil
	}
	if len(char) == 0 {
		char = append(char, ",")
	}
	arr := strings.Split(src, char[0])
	return Cover(arr, true)
}

func Multiple(src string, char1, char2 string) (r [][]int32) {
	if src == "" {
		return
	}
	arr := strings.Split(src, char1)
	for _, s := range arr {
		if v := SplitInt32(s, char2); len(v) > 0 {
			r = append(r, v)
		}
	}
	return r
}
