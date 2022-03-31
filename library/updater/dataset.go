package updater

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
)

type Data bson.M

func (d Data) Has(key string) bool {
	_, ok := d[key]
	return ok
}

func (d Data) Get(key string) interface{} {
	return d[key]
}

func (d Data) Set(key string, val interface{}) interface{} {
	d[key] = val
	return val
}

func (d Data) Add(key string, val int64) (r int64) {
	r = d.GetInt(key) + val
	d[key] = r
	return
}

func (d Data) Sub(key string, val int64) (r int64) {
	r = d.GetInt(key) - val
	d[key] = r
	return
}

func (d Data) GetInt(key string) int64 {
	v, ok := d[key]
	if !ok {
		return 0
	}
	switch v.(type) {
	case int:
		return int64(v.(int))
	case int32:
		return int64(v.(int32))
	case int64:
		return v.(int64)
	case float64:
		return int64(v.(float64))
	case string:
		temp, _ := strconv.ParseInt(v.(string), 10, 64)
		return temp
	default:
		return 0
	}
}

func (d Data) GetInt32(key string) int32 {
	return int32(d.GetInt(key))
}
func (d Data) GetFloat(key string) (r float64) {
	v, ok := d[key]
	if !ok {
		return 0
	}
	switch v.(type) {
	case int:
		r = float64(v.(int))
	case int32:
		r = float64(v.(int32))
	case int64:
		r = float64(v.(int64))
	case float32:
		r = float64(v.(float32))
	case float64:
		r = v.(float64)
	case string:
		r, _ = strconv.ParseFloat(v.(string), 10)
	}
	return
}
func (d Data) GetString(key string) (r string) {
	v, ok := d[key]
	if !ok {
		return ""
	}
	switch v.(type) {
	case string:
		r = v.(string)
	default:
		r = fmt.Sprintf("%v", v)
	}
	return
}

type Dataset struct {
	dataset map[string]itemHMap //oid --> v
	indexes map[int32][]string  //iid --> []oid
}

//Get 使用oid获取道具
func (this *Dataset) Get(id interface{}) (r itemHMap) {
	var iid int32
	var oid string
	switch id.(type) {
	case string:
		oid = id.(string)
	case int32:
		iid = id.(int32)
	case int:
		iid = int32(id.(int))
	case int64:
		iid = int32(id.(int64))
	}
	if iid != 0 && len(this.indexes[iid]) > 0 {
		oid = this.indexes[iid][0]
	}
	if oid != "" {
		r, _ = this.dataset[oid]
	}
	return
}

func (this *Dataset) Set(item itemHMap) {
	iid := item.GetIId()
	oid := item.GetOId()
	if _, ok := this.dataset[oid]; !ok {
		this.indexes[iid] = append(this.indexes[iid], oid)
	}
	this.dataset[oid] = item
}

func (this *Dataset) Del(oid string) {
	item, ok := this.dataset[oid]
	if !ok {
		return
	}
	iid := item.GetIId()
	delete(this.dataset, oid)
	indexes := this.indexes[iid]
	newIndexes := make([]string, 0, len(indexes)-1)
	for _, v := range indexes {
		if v != oid {
			newIndexes = append(newIndexes, v)
		}
	}
	this.indexes[iid] = newIndexes
}

//Val 统计道具数量
func (this *Dataset) Val(iid int32) (r int64) {
	for _, oid := range this.Indexes(iid) {
		if v, ok := this.dataset[oid]; ok {
			r += v.GetVal()
		}
	}
	return
}

//Indexes 配置ID为id的道具oid集合
func (this *Dataset) Indexes(iid int32) (r []string) {
	if v, ok := this.indexes[iid]; ok {
		r = append(r, v...)
	}
	return
}

//release 重置清空数据
func (this *Dataset) release() {
	this.dataset = make(map[string]itemHMap)
	this.indexes = make(map[int32][]string)
}
