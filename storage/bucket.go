package storage

import (
	"runtime/debug"
	"sync"

	"github.com/hwcer/cosgo/uuid"
	"github.com/hwcer/logger"
)

// NewBucket 创建一个新的存储桶
// id: 桶的唯一标识
// cap: 桶的容量
func NewBucket(id int, cap int) *Bucket {
	b := &Bucket{
		dirty:     newDirty(cap),
		values:    make([]Setter, cap),
		Builder:   uuid.New(uint64(id), 1000),
		NewSetter: NewSetterDefault,
	}
	return b
}

// Bucket 一维数组存储器
type Bucket struct {
	dirty     *dirty
	values    []Setter
	Builder   *uuid.Builder
	NewSetter NewSetter    //创建新数据结构
	mu        sync.RWMutex // 读写锁，保护values数组的并发访问
}

// createSocketId 使用index生成ID
func (this *Bucket) createId(index int) string {
	return this.Builder.New(uint64(index)).String(uuid.BaseSize)
}

// parseSocketId 返回idPack中的index
func (this *Bucket) parseId(id string) (int, error) {
	if i, _, err := uuid.Split(id, uuid.BaseSize, 1); err != nil {
		return 0, err
	} else {
		return int(i), nil
	}
}

func (this *Bucket) get(id string) (Setter, bool) {
	index, err := this.parseId(id)
	if err != nil {
		return nil, false
	}
	if index < 0 || index >= len(this.values) || this.values[index] == nil {
		return nil, false
	}
	r := this.values[index]
	if r.Id() != id {
		return nil, false
	}
	return r, true
}

func (this *Bucket) push(v any) Setter {
	const maxAttempts = 100
	this.mu.Lock() // 写入操作加写锁
	defer this.mu.Unlock()
	for i := 0; i < maxAttempts; i++ {
		index := this.dirty.Acquire()
		if index < 0 {
			return nil // 无法获取到可用索引
		}
		if this.values[index] == nil {
			id := this.createId(index)
			setter := this.NewSetter(id, v)
			this.values[index] = setter
			return setter
		}
		// 位置被占用，释放索引回空闲列表
		// 原因：当多个goroutine并发操作时，可能会同时获取到同一个索引
		// 或者在获取索引和检查位置之间，该位置被其他操作占用
		// 释放索引确保空闲列表不会耗尽，资源能够被正确重用
		this.dirty.Release(index)
	}
	return nil
}

func (this *Bucket) Get(id string) (Setter, bool) {
	return this.get(id)
}

func (this *Bucket) Set(id string, v any) bool {
	// 对于Bucket来说，Set是读操作，因为它只是读取并修改values数组中已存在的元素
	// 业务层面的并发安全由业务逻辑来管理，Bucket不负责
	index, err := this.parseId(id)
	if err != nil {
		return false
	}
	if index < 0 || index >= len(this.values) || this.values[index] == nil {
		return false
	}
	setter := this.values[index]
	if setter.Id() != id {
		return false
	}
	setter.Set(v)
	return true
}

// Size 当前数量
func (this *Bucket) Size() int {
	return this.dirty.Size()
}
func (this *Bucket) Free() int {
	return this.dirty.Free()
}

// Range 遍历
func (this *Bucket) Range(f func(Setter) bool) bool {
	for _, val := range this.values {
		if val != nil {
			if !f(val) {
				return false
			}
		}
	}
	return true
}

// Create 创建一个新数据
func (this *Bucket) Create(v any) Setter {
	return this.push(v)
}

// Delete 删除并返回已删除的数据
func (this *Bucket) Delete(id string) Setter {
	index, err := this.parseId(id)
	if err != nil {
		return nil
	}
	this.mu.Lock()         // 写入操作加写锁
	defer this.mu.Unlock() // 释放写锁
	if index < 0 || index >= len(this.values) {
		return nil
	}
	val := this.values[index]
	if val == nil {
		logger.Alert("Bucket Delete error, index:%d,Stack:%s", index, string(debug.Stack()))
		return nil
	}
	if val.Id() != id {
		return nil
	}
	this.values[index] = nil
	this.dirty.Release(index)
	return val
}
