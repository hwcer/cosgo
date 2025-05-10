package utils

import (
	"golang.org/x/exp/constraints"
	"math"
)

// Ceil 除法向上取整
func Ceil[T constraints.Integer](a, b T) (r T) {
	r = a / b
	if a%b != 0 {
		r += 1
	}
	return r
}

func Min[T constraints.Integer](nums ...T) (r T) {
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
func Max[T constraints.Integer](nums ...T) (r T) {
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

// FloatPrecision 四舍五入，保留到Precision位小数
func FloatPrecision(value float64, precision float64) float64 {
	x := math.Pow(10, precision)
	return math.Round(value*x) / x
}
