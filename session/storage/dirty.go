package storage

import (
	"github.com/hwcer/cosgo/logger"
	"runtime/debug"
	"sync"
)

func newDirty(cap int) *dirty {
	d := &dirty{
		list:  make([]int, cap),
		index: -1,
	}
	for i := 0; i < len(d.list); i++ {
		d.list[i] = i
	}
	return d
}

// dirty 已经被删除的index
type dirty struct {
	mu    sync.Mutex
	list  []int
	index int32
}

// Free 当前空余
func (this *dirty) Free() int {
	return len(this.list) - this.Size()
}

// Size 当前已使用
func (this *dirty) Size() int {
	return int(this.index + 1)
}

// Acquire 申请Key
func (this *dirty) Acquire() int {
	if this.Free() <= 0 {
		return -1
	}
	this.mu.Lock()
	defer this.mu.Unlock()
	if this.Free() <= 0 {
		return -1
	}
	this.index++
	val := this.list[this.index]
	this.list[this.index] = -1
	return val
}

// Release 释放Key
func (this *dirty) Release(val int) {
	this.mu.Lock()
	defer this.mu.Unlock()

	if val < 0 || val >= len(this.list) {
		logger.Alert("Bucket dirty Release val error, val:%d,Stack:%s", val, string(debug.Stack()))
		return
	}
	if this.list[this.index] >= 0 {
		logger.Alert("Bucket dirty Release index error, index:%d,Stack:%s", this.index, string(debug.Stack()))
		return
	}
	this.list[this.index] = val
	this.index--
}
