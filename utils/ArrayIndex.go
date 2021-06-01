package utils

func newArrayIndex(cap int) *arrayIndex {
	return &arrayIndex{
		list:  make([]int, 0, cap),
		index: -1,
	}
}

//已经被删除的index
type arrayIndex struct {
	list  []int
	index int
}

func (this *arrayIndex) Add(val int) {
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

func (this *arrayIndex) Get() int {
	if this.index < 0 {
		return -1
	}
	val := this.list[this.index]
	this.list[this.index] = -1
	this.index -= 1
	return val
}

func (this *arrayIndex) Size() int {
	return this.index + 1
}
