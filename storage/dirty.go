package storage

import (
	"github.com/hwcer/logger"
)

// newDirty 创建一个新的 dirty 实例
// cap: 槽位总数
func newDirty(cap int) *dirty {
	d := &dirty{
		list: make([]int, cap),
		free: make([]bool, cap),
	}
	for i := 0; i < cap; i++ {
		d.list[i] = i
		d.free[i] = true
	}
	return d
}

// dirty 空闲槽位索���栈���LIFO）
// 通过 free 位图��证每个索��在栈中唯一，防止双重释放
// 注意：dirty 不包含���部锁，依赖外部 Bucket.mu 保护
type dirty struct {
	list []int  // 空闲索引栈，从末尾弹出/压入
	free []bool // free[i]=true 表示槽位 i 在空闲栈中（未被占用）
}

// Free 当前空闲槽位数
func (this *dirty) Free() int {
	return len(this.list)
}

// Size 当���已使用槽位数
func (this *dirty) Size() int {
	return cap(this.list) - len(this.list)
}

// Acquire 从栈顶弹出一个空闲索引，O(1)
// ���回 -1 表示无可用槽位
func (this *dirty) Acquire() int {
	if len(this.list) == 0 {
		return -1
	}
	lastIndex := len(this.list) - 1
	val := this.list[lastIndex]
	this.list = this.list[:lastIndex]
	this.free[val] = false
	return val
}

// Release 将索引归还到空闲栈，O(1)
// 内置双重释放检测：如果索引已在空闲栈中，记录告警并忽略
func (this *dirty) Release(val int) {
	if val < 0 || val >= len(this.free) {
		logger.Alert("dirty Release: invalid index %d", val)
		return
	}
	if this.free[val] {
		logger.Alert("dirty Release: double release of index %d", val)
		return
	}
	this.free[val] = true
	this.list = append(this.list, val)
}
