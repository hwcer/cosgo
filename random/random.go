package utils

type Random struct {
	items map[int32]int32
	limit int32
}

func New(items map[int32]int32) *Random {
	i := &Random{}
	i.items = map[int32]int32{}
	for k, v := range items {
		i.items[k] = v
		i.limit += v
	}
	return i
}

func (this *Random) Roll() int32 {
	if this.limit == 0 {
		return -1
	}
	rnd := int32(Roll(1, int(this.limit)))
	for i, v := range this.items {
		if v > 0 {
			rnd -= v
			if rnd <= 0 {
				return i
			}
		}
	}
	return -1
}

// Multi 随机多个不重复
func (this *Random) Multi(num int) (r []int32) {
	if num >= len(this.items) {
		for k, _ := range this.items {
			r = append(r, k)
		}
		return
	}
	items := make(map[int32]int32, len(this.items))
	limit := this.limit
	for k, v := range this.items {
		items[k] = v
	}
	for i := 0; i < num; i++ {
		rnd := int32(Roll(1, int(limit)))
		for j, v := range this.items {
			rnd -= v
			if rnd <= 0 {
				r = append(r, j)
				delete(items, j)
				limit -= v
				break
			}
		}
	}

	return
}
