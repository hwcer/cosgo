package utils

import (
	"math"
)

// Integer 约束所有整数类型(替代golang.org/x/exp/constraints.Integer)
type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Ceil 除法向上取整
func Ceil[T Integer](a, b T) (r T) {
	r = a / b
	if a%b != 0 {
		r += 1
	}
	return r
}

func Min[T Integer](nums ...T) (r T) {
	if len(nums) == 0 {
		return
	}
	r = nums[0]
	for _, num := range nums[1:] {
		r = min(r, num)
	}
	return r
}
func Max[T Integer](nums ...T) (r T) {
	if len(nums) == 0 {
		return
	}
	r = nums[0]
	for _, num := range nums[1:] {
		r = max(r, num)
	}
	return r
}

// FloatPrecision 四舍五入，保留到Precision位小数
func FloatPrecision(value float64, precision float64) float64 {
	x := math.Pow(10, precision)
	return math.Round(value*x) / x
}
