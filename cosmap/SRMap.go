package cosmap

import (
	"sync"
)

//SRMap SafeReadMap读写安全，读多写少情况表现极好

func NewSRMap(cap int) *SRMap {
	return &SRMap{
		fields: make(map[string]int, cap),
		values: make([]interface{}, 0, cap),
		remove: NewArrayIndex(cap),
	}
}

type SRMap struct {
	mutex  sync.Mutex
	remove *ArrayIndex
	fields map[string]int
	values []interface{}
}

func (this *SRMap) Set(key string, val interface{}) {
	if index := this.indexOf(key); index >= 0 {
		this.values[index] = val
	} else {
		this.MSet(map[string]interface{}{key: val})
	}
}

func (this *SRMap) Get(key string) (interface{}, bool) {
	if index := this.indexOf(key); index >= 0 {
		return this.values[index], true
	} else {
		return nil, false
	}
}

//批量添加
func (this *SRMap) MSet(data map[string]interface{}) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	fields := make(map[string]int, len(this.fields)+len(data))
	for k, v := range data {
		if index, ok := this.appendValue(k, v); ok {
			fields[k] = index
		}
	}
	this.appendField(fields)
}

func (this *SRMap) Delete(key string) bool {
	index := this.indexOf(key)
	if index < 0 {
		return false
	}
	this.mutex.Lock()
	defer this.mutex.Unlock()
	delete(this.fields, key)
	this.values[index] = nil
	this.remove.Add(index)
	return true
}

//遍历
func (this *SRMap) Range(f func(k string, v interface{})) {
	for k, i := range this.fields {
		f(k, this.values[i])
	}
}

func (this *SRMap) indexOf(key string) int {
	if index, ok := this.fields[key]; ok {
		return index
	}
	return -1
}

//加锁前提下执行
func (this *SRMap) appendValue(key string, val interface{}) (index int, new bool) {
	if index = this.indexOf(key); index >= 0 {
		this.values[index] = val
	} else if index = this.remove.Get(); index >= 0 {
		new = true
		this.fields[key] = index
		this.values[index] = val
	} else {
		new = true
		index = len(this.values)
		this.values = append(this.values, val)
	}
	return
}

//加锁前提下执行
func (this *SRMap) appendField(fields map[string]int) {
	if len(fields) == 0 {
		return
	}
	for k, v := range this.fields {
		fields[k] = v
	}
	this.fields = fields
}
