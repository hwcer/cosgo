package random

import "math/rand"

func Roll(a, b int) int {
	if b <= a {
		return a
	}
	return a + rand.Intn(b-a)
}

// Probability 独立概率，args单位，默认万分比，，，百分比传入100
func Probability(per int, args ...int) bool {
	if per <= 0 {
		return false
	}
	var unit = 10000
	if len(args) > 0 && args[0] != 0 {
		unit = args[0]
	}
	if per >= unit {
		return true
	}
	return per >= Roll(1, unit)
}

// Relative 相对概率，权重
func Relative(items map[int32]int32) int32 {
	l := len(items)
	if l == 0 {
		return -1
	} else if l == 1 {
		for k, _ := range items {
			return k
		}
	}
	var total int32 = 0
	for _, v := range items {
		if v > 0 {
			total += v
		}
	}
	if total == 0 {
		return -1
	}
	rnd := int32(Roll(1, int(total)))
	for i, v := range items {
		if v > 0 {
			rnd -= v
			if rnd <= 0 {
				return i
			}
		}
	}
	return -1
}

// RelativeMulti 相对概率，权重 返回多个,repeat 是否可以重复
func RelativeMulti(items map[int32]int32, num int, repeat ...bool) []int32 {
	var total int32 = 0
	for _, v := range items {
		if v > 0 {
			total += v
		}
	}
	if total == 0 {
		return nil
	}

	re := false
	if len(repeat) > 0 && repeat[0] == true {
		re = true
	}
	ret := make([]int32, num)
	for i := 0; i < num; i++ {
		rnd := int32(Roll(1, int(total)))
		for it, v := range items {
			if v > 0 {
				rnd -= v
				if rnd <= 0 {
					ret[i] = it
					if !re {
						total -= v
						delete(items, it)
					}
					break
				}
			}
		}
	}
	return ret
}
