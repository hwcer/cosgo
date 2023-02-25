package utils

import (
	"math/rand"
	"time"
)

var Random *random

func init() {
	Random = NewRandom("0123456789abcdefghijklmnopqrstuvwxyz")
}

type random struct {
	rand  *rand.Rand
	bytes []byte
}

func NewRandom(randomStringSeed string) *random {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &random{bytes: []byte(randomStringSeed), rand: rand}
}

func (this *random) Roll(a, b int) int {
	if b <= a {
		return a
	}
	return a + this.rand.Intn(b-a)
}

func (this *random) String(l int) string {
	result := make([]byte, l)
	m := len(this.bytes) - 1
	for i := 0; i < l; i++ {
		result[i] = this.bytes[this.Roll(0, m)]
	}
	return string(result)
}

// Probability 独立概率，args单位，默认万分比，，，百分比传入100
func (this *random) Probability(per int, args ...int) bool {
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
	return per >= this.Roll(1, unit)
}

// Relative 相对概率，权重
func (this *random) Relative(items map[int32]int32) int32 {
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
	rnd := int32(Random.Roll(1, int(total)))
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
func (this *random) RelativeMulti(items map[int32]int32, num int, repeat ...bool) []int32 {
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
		rnd := int32(Random.Roll(1, int(total)))
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
