package random

import (
	"sort"
)

type Random struct {
	items [][2]int32 //[id,weight]
	total int32
}

func New(items map[int32]int32) *Random {
	r := &Random{}
	for k, v := range items {
		r.items = append(r.items, [2]int32{k, v})
		r.total += v
	}
	sort.Slice(r.items, func(i, j int) bool {
		return r.items[i][1] > r.items[j][1]
	})
	return r
}

// Roll 简单的权重算法，直接计算区间落点
func (this *Random) Roll() int32 {
	if this.total == 0 {
		return -1
	}
	rnd := Roll(1, this.total)
	for _, v := range this.items {
		if v[1] > 0 {
			rnd -= v[1]
			if rnd <= 0 {
				return v[0]
			}
		}
	}
	return -1
}

// Weight 权重算法,每一条执行一次随机，如果不命中继续下一条
// 按照权重从小到大执行, 最后一条(权重最大)作为保底
// 执行结果和 Roll 基本一致
// 优点是极限值出现的时机更加靠后,策划更加满意和放心
// 缺点是更加消耗硬件
func (this *Random) Weight() (r int32) {
	if this.total == 0 {
		return -1
	}
	for i := len(this.items) - 1; i >= 0; i-- {
		v := this.items[i]
		r = v[0]
		if rnd := Roll(1, this.total); v[1] >= rnd {
			return
		}
	}
	return
}

// Multi 随机多个不重复
func (this *Random) Multi(num int) (r []int32) {
	if num >= len(this.items) {
		for _, v := range this.items {
			r = append(r, v[0])
		}
		return
	}
	items := make([][2]int32, len(this.items))
	limit := this.total
	copy(items, this.items)

	for i := 0; i < num; i++ {
		rnd := Roll(1, limit)
		for j, v := range items {
			if v[1] <= 0 {
				continue
			}
			rnd -= v[1]
			if rnd <= 0 {
				r = append(r, v[0])
				limit -= v[1]
				items[j] = [2]int32{v[0], 0}
				break
			}
		}
	}
	return
}

func (this *Random) Range(f func(k, v int32) bool) {
	for _, v := range this.items {
		if !f(v[0], v[1]) {
			return
		}
	}
}
