package cache

import (
	"math"
	"sync"
)

func New(cap int) *Cache {
	return &Cache{
		seed:   seedDefaultValue,
		dirty:  newDirty(cap),
		values: make([]Dataset, cap, cap),
	}
}

// Cache 一维数组存储器，，读，修改
type Cache struct {
	seed      uint32 //Index 生成种子
	mutex     sync.Mutex
	dirty     *dirty
	values    []Dataset
	NewSetter NewDataFunc //创建新数据结构
}

//createSocketId 使用index生成ID
func (this *Cache) createId(index int) uint64 {
	this.seed++
	if this.seed >= math.MaxUint32 {
		this.seed = seedDefaultValue
	}
	return uint64(this.seed)<<32 + uint64(index)
}

//parseSocketId 返回idPack中的index
func (this *Cache) parseId(id uint64) int {
	if id == 0 {
		return -1
	}
	return int(id & 0xffffffff)
}
func (this *Cache) new(id uint64, val interface{}) (data Dataset) {
	if this.NewSetter != nil {
		data = this.NewSetter(id, val)
	} else {
		data = NewData()
	}
	if data.Id() == 0 {
		data.Reset(id, data.Get())
	}
	return
}

func (this *Cache) get(id uint64) (Dataset, bool) {
	index := this.parseId(id)
	if index < 0 || index >= len(this.values) || this.values[index] == nil || this.values[index].Id() != id {
		return nil, false
	}
	return this.values[index], true
}

func (this *Cache) set(index int, val interface{}) uint64 {
	size := len(this.values)
	if index < 0 || index > size {
		index = size
	}
	id := this.createId(index)

	if index == size {
		this.values = append(this.values, this.new(id, val)) //扩容
	} else if this.values[index] == nil {
		this.values[index] = this.new(id, val)
	} else {
		this.values[index].Reset(id, val)
	}
	return id
}

func (this *Cache) remove(id uint64) Dataset {
	index := this.parseId(id)
	if index < 0 || index >= len(this.values) || this.values[index].Id() != id {
		return nil
	}
	val := this.values[index]
	this.values[index] = nil
	this.dirty.Release(index)
	return val
}

func (this *Cache) Get(id uint64) (Dataset, bool) {
	return this.get(id)
}

func (this *Cache) Set(id uint64, v interface{}) bool {
	updater, ok := this.get(id)
	if !ok {
		return ok
	}
	updater.Set(v)
	return true
}

//Push 放入一个新数据
func (this *Cache) Push(v interface{}) uint64 {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if index := this.dirty.Acquire(); index >= 0 && index < len(this.values) && (this.values[index] == nil || this.values[index].Id() == 0) {
		return this.set(index, v)
	}
	id := this.set(-1, v)
	return id
}

//Size 当前数量
func (this *Cache) Size() int {
	return len(this.values) - this.dirty.Size()
}

//Range 遍历
func (this *Cache) Range(f func(Dataset) bool) {
	for _, val := range this.values {
		if val != nil && val.Id() > 0 && !f(val) {
			break
		}
	}
}

//Remove 批量移除
func (this *Cache) Remove(ids ...uint64) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	for _, id := range ids {
		this.remove(id)
	}
	return
}

//Delete 删除并返回已删除的数据
func (this *Cache) Delete(id uint64) Dataset {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.remove(id)
}
