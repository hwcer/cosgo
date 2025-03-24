package storage

import (
	"github.com/hwcer/cosgo/logger"
	"runtime/debug"
	"sync/atomic"
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
	i := atomic.AddInt32(&this.index, 1)
	if i >= int32(len(this.list)) {
		atomic.AddInt32(&this.index, -1)
		return -1
	}
	val := this.list[i]
	this.list[i] = -1
	return val
}

// Release 释放Key
func (this *dirty) Release(val int) {
	if val < 0 || val >= len(this.list) {
		logger.Alert("Bucket dirty Release val error, val:%d,Stack:%s", val, string(debug.Stack()))
		return
	}
	i := atomic.AddInt32(&this.index, -1)
	if i < 0 {
		atomic.AddInt32(&this.index, 1)
		logger.Alert("Bucket dirty Release index error, index:%d,Stack:%s", i, string(debug.Stack()))
		return
	}
	this.list[i] = val
}
