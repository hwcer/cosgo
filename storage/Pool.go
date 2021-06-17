package storage

import "sync"

func NewPool(cap int, init bool) *Pool {
	pool := &Pool{
		list:  make([]int, 0, cap),
		index: -1,
	}
	if !init {
		return pool
	}
	for i := 0; i < cap; i++ {
		pool.release(i)
	}

	return pool
}

//Pool 已经被删除的index
type Pool struct {
	mu    sync.Mutex
	list  []int
	index int
}

func (this *Pool) Size() int {
	return this.index + 1
}

//申请Key
func (this *Pool) Acquire() int {
	if this.index < 0 {
		return -1
	}
	this.mu.Lock()
	val := this.list[this.index]
	this.list[this.index] = -1
	this.index -= 1
	this.mu.Unlock()
	return val
}

//释放Key
func (this *Pool) Release(val int) {
	if val < 0 {
		return
	}
	this.mu.Lock()
	this.release(val)
	this.mu.Unlock()
}

func (this *Pool) release(val int) {
	this.index += 1
	if this.index < len(this.list) {
		this.list[this.index] = val
	} else {
		this.list = append(this.list, val)
	}
}
