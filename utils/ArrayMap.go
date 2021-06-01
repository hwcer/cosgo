package utils

import (
	"sync"
)

type ArrayMapKey interface {
	Equal(interface{}) bool //比较2个KEY是否相等
}

func NewArrayMap(cap int) *ArrayMap {
	return &ArrayMap{
		keys:   make([]ArrayMapKey, 0, cap),
		values: make([]interface{}, 0, cap),
		remove: newArrayIndex(cap),
	}
}

type ArrayMap struct {
	keys   []ArrayMapKey
	values []interface{}
	mutex  sync.Mutex
	remove *arrayIndex
}

func (this *ArrayMap) Set(key ArrayMapKey, val interface{}) {
	if index := this.indexOf(key); index >= 0 {
		this.values[index] = val
	} else {
		this.append(key, val)
	}
}

func (this *ArrayMap) Get(key ArrayMapKey) (interface{}, bool) {
	if index := this.indexOf(key); index >= 0 {
		return this.values[index], true
	} else {
		return nil, false
	}
}

func (this *ArrayMap) Delete(key ArrayMapKey) bool {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if index := this.indexOf(key); index >= 0 {
		this.keys[index] = nil
		this.values[index] = nil
		this.remove.Add(index)
		return true
	}
	return false
}

//遍历
func (this *ArrayMap) Range(f func(k ArrayMapKey, v interface{})) {
	for i, k := range this.keys {
		if k != nil {
			f(k, this.values[i])
		}
	}
}

func (this *ArrayMap) indexOf(key ArrayMapKey) int {
	for i, k := range this.keys {
		if k.Equal(key) {
			return i
		}
	}
	return -1
}
func (this *ArrayMap) append(key ArrayMapKey, val interface{}) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if index := this.indexOf(key); index >= 0 {
		this.values[index] = val
	} else if index = this.remove.Get(); index >= 0 && this.keys[index] == nil {
		this.keys[index] = key
		this.values[index] = val
	} else {
		this.keys = append(this.keys, key)
		this.values = append(this.values, val)
	}
}
