package storage

import (
	"runtime/debug"

	"github.com/hwcer/cosgo/uuid"
	"github.com/hwcer/logger"
)

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
	NewSetter NewSetter //创建新数据结构
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
	index := this.dirty.Acquire()
	if index < 0 {
		return nil
	}
	if s := this.values[index]; s != nil {
		return this.push(v)
	}
	id := this.createId(index)
	setter := this.NewSetter(id, v)
	this.values[index] = setter
	return setter
}

func (this *Bucket) Get(id string) (Setter, bool) {
	return this.get(id)
}

func (this *Bucket) Set(id string, v any) bool {
	setter, ok := this.get(id)
	if ok {
		setter.Set(v)
	}
	return ok
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
