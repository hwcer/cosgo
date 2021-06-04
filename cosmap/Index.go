package cosmap

func NewArrayIndex(cap int) *ArrayIndex {
	return &ArrayIndex{
		list:  make([]int, 0, cap),
		index: -1,
	}
}

//已经被删除的index
type ArrayIndex struct {
	list  []int
	index int
}

func (this *ArrayIndex) Add(val int) {
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

func (this *ArrayIndex) Get() int {
	if this.index < 0 {
		return -1
	}
	val := this.list[this.index]
	this.list[this.index] = -1
	this.index -= 1
	return val
}

func (this *ArrayIndex) Size() int {
	return this.index + 1
}
