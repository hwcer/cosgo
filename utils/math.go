package utils

import "golang.org/x/exp/constraints"

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
