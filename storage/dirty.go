package storage

func newDirty(cap int) *dirty {
	d := &dirty{
		list:  make([]int, cap),
		index: -1,
	}
	for i := cap - 1; i >= 0; i-- {
		d.Release(i)
	}
	return d
}

// dirty 已经被删除的index
type dirty struct {
	list  []int
	index int
}

func (this *dirty) Size() int {
	return this.index + 1
}

// 申请Key
func (this *dirty) Acquire() int {
	if this.index < 0 {
		return -1
	}
	val := this.list[this.index]
	this.list[this.index] = -1
	this.index -= 1
	return val
}

// 释放Key
func (this *dirty) Release(val int) {
	if val < 0 {
		return
	}
	this.index += 1
	if this.index < len(this.list) {
		this.list[this.index] = val
	} else {
		this.list = append(this.list, val)
	}
}
