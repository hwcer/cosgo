package storage

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/hwcer/logger"
)

// entry 包装 Setter 接口，使其可以通过单个 unsafe.Pointer 原子读写
// 每次 New() 创建新 entry（不复用），保证读端永远看到完整的 Setter 接口值
type entry struct {
	setter Setter
}

// NewBucket 创建一个新的存储桶
// id: 桶索引（uint16，编码进 token 前 4 个 hex 字符）
// cap: 桶的容量（预分配，生命周期内不变）
func NewBucket(id int, cap int) *Bucket {
	b := &Bucket{
		id:        uint16(id),
		dirty:     newDirty(cap),
		values:    make([]unsafe.Pointer, cap),
		NewSetter: NewSetterDefault,
	}
	// 初始化 pos 为满，迫使首次 tokenEncode 调用 crypto/rand.Read 填充缓冲区
	b.rng.pos = randBufSize
	return b
}

// Bucket 一维数组存储器
//
// 并发模型：
//   - values 切片预分配固定长度，生命周期内不变，slice header 无 race
//   - 每个槽位通过 atomic.LoadPointer/StorePointer 原子读写，无 torn read
//   - 写路径（New / Delete）使用 sync.Mutex 串行化
//   - 读路径（Get / Range）完全无锁
//   - 禁止在 Range 回调中调用 New/Delete
type Bucket struct {
	id        uint16           // 桶索引，编码进 token
	dirty     *dirty           // 空闲槽位管理
	values    []unsafe.Pointer // 每个元素是 *entry，原子操作
	rng       randBuffer       // 随机数缓冲，在 mu.Lock 下使用
	NewSetter NewSetter        // 创建新 Setter 的回调
	mu        sync.Mutex       // 保护写路径
	size      atomic.Int32     // 已占用槽位数
}

// loadSlot 原子读取指定槽位的 Setter，无锁
func (this *Bucket) loadSlot(index int) Setter {
	p := atomic.LoadPointer(&this.values[index])
	if p == nil {
		return nil
	}
	return (*entry)(p).setter
}

// storeSlot 原子写入指定槽位（在 mu.Lock 下调用）
// 每次写入创建新 entry，保证读端看到的旧 entry 不会被修改
func (this *Bucket) storeSlot(index int, s Setter) {
	if s == nil {
		atomic.StorePointer(&this.values[index], nil)
	} else {
		atomic.StorePointer(&this.values[index], unsafe.Pointer(&entry{setter: s}))
	}
}

// Get 原子读：解析 token 得到 slot 索引，直接数组定位，全串比较校验
// 零分配，~12ns
func (this *Bucket) Get(id string) (Setter, bool) {
	index, ok := tokenDecodeSlot(id)
	if !ok || index < 0 || index >= len(this.values) {
		return nil, false
	}
	r := this.loadSlot(index)
	if r == nil || r.Id() != id {
		return nil, false
	}
	return r, true
}

// Size 当前已占用槽位数，O(1) 无锁
func (this *Bucket) Size() int {
	return int(this.size.Load())
}

// Free 当前空闲槽位数，O(1) 无锁
func (this *Bucket) Free() int {
	return len(this.values) - int(this.size.Load())
}

// Range 无锁遍历，通过 loadSlot 原子读取每个槽位
// 禁止在回调中调用 New/Delete
func (this *Bucket) Range(f func(Setter) bool) bool {
	for i := range this.values {
		if val := this.loadSlot(i); val != nil {
			if !f(val) {
				return false
			}
		}
	}
	return true
}

// New 分配一个新槽位并写入对象
// token 生成使用 crypto/rand 缓冲池（摊薄后 ~3ns），整体仅 3 次堆分配
func (this *Bucket) New(v any) Setter {
	this.mu.Lock()
	defer this.mu.Unlock()
	index := this.dirty.Acquire()
	if index < 0 {
		return nil
	}
	id := tokenEncode(this.id, uint32(index), &this.rng)
	setter := this.NewSetter(id, v)
	this.storeSlot(index, setter)
	this.size.Add(1)
	return setter
}

// Delete 释放槽位并返回被删除的对象
// 对同一 ID 多次 Delete 是安全的（第二次返回 nil）
func (this *Bucket) Delete(id string) Setter {
	index, ok := tokenDecodeSlot(id)
	if !ok {
		return nil
	}
	this.mu.Lock()
	defer this.mu.Unlock()
	if index < 0 || index >= len(this.values) {
		return nil
	}
	val := this.loadSlot(index)
	if val == nil {
		logger.Alert("Bucket Delete: slot already empty, index:%d", index)
		return nil
	}
	if val.Id() != id {
		return nil
	}
	this.storeSlot(index, nil)
	this.dirty.Release(index)
	this.size.Add(-1)
	return val
}
