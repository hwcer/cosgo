package random

import (
	"sort"
)

type Data struct {
	Key int32 //序号
	Val int32 //权重
}

func (this *Data) GetKey() int32 {
	return this.Key
}
func (this *Data) GetVal() int32 {
	return this.Val
}

type Less func(a, b *Data) bool

type Random struct {
	less  Less
	items []*Data //[id,Roll]
	total int32
}

func New(items map[int32]int32, less ...Less) *Random {
	r := &Random{}
	for k, v := range items {
		r.items = append(r.items, &Data{Key: k, Val: v})
		r.total += v
	}
	if len(less) > 0 {
		r.less = less[0]
	} else {
		r.less = func(i, j *Data) bool { return i.Val > j.Val }
	}
	r.Sort(r.less)
	return r
}

func (this *Random) Sort(f Less) {
	sort.Slice(this.items, func(i, j int) bool {
		return f(this.items[i], this.items[j])
	})
}

// Roll 简单的权重算法，直接计算区间落点
func (this *Random) Roll() int32 {
	if this.total == 0 {
		return -1
	}
	rnd := Roll(1, this.total)
	for _, d := range this.items {
		if d.Val > 0 {
			rnd -= d.Val
			if rnd <= 0 {
				return d.Key
			}
		}
	}
	return -1
}
func (this *Random) Weight() (r int32) {
	if this.total == 0 {
		return -1
	}
	for _, v := range this.items {
		r = v.GetKey()
		if n := Roll(1, this.total); v.GetVal() >= n {
			return
		}
	}
	return
}

// Multi 随机多个不重复
func (this *Random) Multi(num int) (r []int32) {
	if num >= len(this.items) {
		for _, d := range this.items {
			r = append(r, d.Key)
		}
		return
	}
	items := make([]*Data, len(this.items))
	limit := this.total
	copy(items, this.items)

	for i := 0; i < num; i++ {
		rnd := Roll(1, limit)
		for j, d := range items {
			if d.Val <= 0 {
				continue
			}
			rnd -= d.Val
			if rnd <= 0 {
				r = append(r, d.Key)
				limit -= d.Val
				items[j] = &Data{Key: d.Key, Val: 0}
				break
			}
		}
	}
	return
}

func (this *Random) Range(f func(k, v int32) bool) {
	for _, d := range this.items {
		if !f(d.Key, d.Val) {
			return
		}
	}
}

// Filter 根据 filter 剔除不符合规则的项目
func (this *Random) Filter(filter func(k, v int32) bool) *Random {
	items := make(map[int32]int32)
	for _, d := range this.items {
		if filter(d.Key, d.Val) {
			items[d.Key] = d.Val
		}
	}
	return New(items, this.less)
}

// Reset 根据 filter 重新设定权重
func (this *Random) Reset(filter func(k, v int32) int32) *Random {
	items := make(map[int32]int32)
	for _, d := range this.items {
		if n := filter(d.Key, d.Val); n > 0 {
			items[d.Key] = n
		}
	}
	return New(items, this.less)
}
