package storage

import (
	"runtime/debug"

	"github.com/hwcer/logger"
)

// newDirty 创建一个新的dirty实例
// cap: 初始化容量
func newDirty(cap int) *dirty {
	d := &dirty{
		list: make([]int, cap),
	}
	for i := 0; i < len(d.list); i++ {
		d.list[i] = i
	}
	return d
}

// dirty 管理已删除的索引，用于高效重用
// 注意：dirty 不包含内部锁，依赖外部锁保护
// 所有操作必须在外部锁的保护下进行
type dirty struct {
	list []int // 存储空闲索引的切片
}

// Free 当前空余
func (this *dirty) Free() int {
	return len(this.list)
}

// Size 当前已使用
func (this *dirty) Size() int {
	return cap(this.list) - len(this.list)
}

// Acquire 申请Key
func (this *dirty) Acquire() int {
	if len(this.list) == 0 {
		return -1
	}
	// 从切片末尾弹出一个索引
	lastIndex := len(this.list) - 1
	val := this.list[lastIndex]
	this.list = this.list[:lastIndex]
	return val
}

// Release 释放Key
func (this *dirty) Release(val int) {
	if val < 0 || val >= cap(this.list) {
		logger.Alert("Bucket dirty Release val error, val:%d,Stack:%s", val, string(debug.Stack()))
		return
	}
	// 将索引追加到切片末尾
	this.list = append(this.list, val)
}
