package smap

import (
	"math"
	"sync"
)

func newSetterDefault(id uint64, data interface{}) Interface {
	return NewSetter(id, data)
}

func New(cap int) *Array {
	return &Array{
		seed:      seedDefaultValue,
		dirty:     newDirty(cap),
		values:    make([]Interface, cap, cap),
		NewSetter: newSetterDefault,
	}
}

// Array 一维数组存储器，，读，修改
type Array struct {
	seed      uint32     //Index 生成种子
	mutex     sync.Mutex //仅添加时需要锁
	dirty     *dirty
	values    []Interface
	NewSetter func(id uint64, val interface{}) Interface //创建新数据结构
}

//createSocketId 使用index生成ID
func (this *Array) createId(index int) uint64 {
	this.seed++
	if this.seed >= math.MaxUint32 {
		this.seed = seedDefaultValue
	}
	return uint64(this.seed)<<32 + uint64(index)
}

//parseSocketId 返回idPack中的index
func (this *Array) parseId(id uint64) int {
	if id == 0 {
		return -1
	}
	return int(id & 0xffffffff)
}

func (this *Array) get(id uint64) (Interface, bool) {
	index := this.parseId(id)
	if index < 0 || index >= len(this.values) || this.values[index] == nil || this.values[index].Id() != id {
		return nil, false
	}
	return this.values[index], true
}

func (this *Array) set(index int, val interface{}) (setter Interface) {
	size := len(this.values)
	if index < 0 || index > size {
		index = size
	}
	id := this.createId(index)
	setter = this.NewSetter(id, val)
	if index == size {
		this.values = append(this.values, setter) //扩容
	} else if this.values[index] == nil {
		this.values[index] = setter
	}
	return
}

func (this *Array) remove(id uint64) Interface {
	index := this.parseId(id)
	if index < 0 || index >= len(this.values) || this.values[index].Id() != id {
		return nil
	}
	val := this.values[index]
	this.values[index] = nil
	this.dirty.Release(index)
	return val
}

func (this *Array) Get(id uint64) (Interface, bool) {
	return this.get(id)
}

func (this *Array) Set(id uint64, v interface{}) bool {
	setter, ok := this.get(id)
	if !ok {
		return ok
	}
	setter.Set(v)
	return true
}

//Push 放入一个新数据
func (this *Array) Push(v interface{}) Interface {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if index := this.dirty.Acquire(); index >= 0 && index < len(this.values) && (this.values[index] == nil || this.values[index].Id() == 0) {
		return this.set(index, v)
	}
	return this.set(-1, v)
}

//Size 当前数量
func (this *Array) Size() int {
	return len(this.values) - this.dirty.Size()
}

//Range 遍历
func (this *Array) Range(f func(Interface) bool) {
	for _, val := range this.values {
		if val != nil && val.Id() > 0 && !f(val) {
			break
		}
	}
}

//Remove 批量移除
func (this *Array) Remove(ids ...uint64) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	for _, id := range ids {
		this.remove(id)
	}
	return
}

//Delete 删除并返回已删除的数据
func (this *Array) Delete(id uint64) Interface {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.remove(id)
}
