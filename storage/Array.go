package storage

import (
	"sync"
)

type ArrayDataset interface {
	Id() uint64
	Get() interface{}
	Set(interface{})
	reset(id uint64, data interface{})
}

type ArrayDatasetDefault struct {
	id   uint64
	data interface{}
}

type ArrayDatasetHandler func(uint64, interface{}) ArrayDataset

func (this *ArrayDatasetDefault) Id() uint64 {
	return this.id
}
func (this *ArrayDatasetDefault) Get() interface{} {
	return this.data
}
func (this *ArrayDatasetDefault) Set(data interface{}) {
	this.data = data
}

//内部接口
func (this *ArrayDatasetDefault) reset(id uint64, data interface{}) {
	this.id = id
	this.data = data
}

func NewArray(cap int) *Array {
	return &Array{
		seed:       1,
		dirty:      NewPool(cap, true),
		values:     make([]ArrayDataset, cap, cap),
		Multiplex:  true,
		NewDataset: NewArrayDataset,
	}
}

func NewArrayDataset(id uint64, data interface{}) ArrayDataset {
	return &ArrayDatasetDefault{id: id, data: data}
}

// Array 一维数组存储器，，读，修改
type Array struct {
	seed       uint32 //ID 生成种子
	mutex      sync.Mutex
	dirty      *Pool
	values     []ArrayDataset
	Multiplex  bool                //ArrayDataset 是否可以复用，默认true
	NewDataset ArrayDatasetHandler //创建新数据结构
}

//createSocketId 使用index生成ID
func (this *Array) createId(index int) uint64 {
	this.seed++
	return uint64(index)<<32 | uint64(this.seed)
}

//parseSocketId 返回idPack中的index
func (this *Array) parseId(id uint64) int {
	if id == 0 {
		return -1
	}
	return int(id >> 32)
}
func (this *Array) get(id uint64) (ArrayDataset, bool) {
	index := this.parseId(id)
	if index < 0 || index >= len(this.values) || this.values[index] == nil || this.values[index].Id() != id {
		return nil, false
	}
	return this.values[index], true
}

func (this *Array) set(index int, val interface{}) uint64 {
	size := len(this.values)
	if index < 0 || index > size {
		index = size
	}
	id := this.createId(index)
	if index == size {
		this.values = append(this.values, this.NewDataset(id, val)) //扩容
	} else if this.values[index] == nil {
		this.values[index] = this.NewDataset(id, val)
	} else {
		this.values[index].reset(id, val)
	}
	return id
}

//Get 获取
func (this *Array) Get(id uint64) (interface{}, bool) {
	if v, ok := this.get(id); ok {
		return v.Get(), true
	} else {
		return nil, false
	}
}

func (this *Array) Set(v interface{}) uint64 {
	if index := this.dirty.Acquire(); index >= 0 && index < len(this.values) && (this.values[index] == nil || this.values[index].Id() == 0) {
		return this.set(index, v)
	}
	this.mutex.Lock()
	id := this.set(-1, v)
	this.mutex.Unlock()
	return id
}

//Size 当前数量
func (this *Array) Size() int {
	return len(this.values) - this.dirty.Size()
}

//Delete 删除
func (this *Array) Delete(id uint64) bool {
	index := this.parseId(id)
	if index < 0 || index >= len(this.values) || this.values[index].Id() != id {
		return false
	}
	if this.Multiplex {
		this.values[index].reset(0, nil)
	} else {
		this.values[index] = nil
	}
	this.dirty.Release(index)
	return true
}

//Range 遍历
func (this *Array) Range(f func(item ArrayDataset) bool) {
	for _, val := range this.values {
		if val != nil && val.Id() > 0 && !f(val) {
			break
		}
	}
}

//Get 获取ArrayDataset
func (this *Array) Dataset(id uint64) (ArrayDataset, bool) {
	return this.get(id)
}
