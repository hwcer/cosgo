package session

import (
	"github.com/hwcer/cosgo/utils"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const memoryDatasetBitSize = 36

func NewMemoryDataset(data map[string]interface{}) *MemoryDataset {
	d := &MemoryDataset{keys: make([]string, 0, len(data)), values: make([]interface{}, 0, len(data))}
	for k, v := range data {
		d.keys = append(d.keys, k)
		d.values = append(d.values, v)
	}
	d.Reset()
	return d
}

type MemoryDataset struct {
	id     string
	keys   []string
	values []interface{}
	mutex  sync.Mutex
	locked int32
	expire int64
}

//Id 获取session id
func (this *MemoryDataset) Id() string {
	return this.id
}

func (this *MemoryDataset) Set(key string, val interface{}) {
	if index := this.indexOf(key); index >= 0 {
		this.values[index] = val
	} else {
		this.append(key, val)
	}
}

func (this *MemoryDataset) Get(key string) (interface{}, bool) {
	if index := this.indexOf(key); index >= 0 {
		return this.values[index], true
	} else {
		return nil, false
	}
}
func (this *MemoryDataset) Lock() bool {
	return atomic.CompareAndSwapInt32(&this.locked, 0, 1)
}

func (this *MemoryDataset) Reset() {
	if Options.MaxAge > 0 {
		this.expire = time.Now().Unix() + Options.MaxAge
	}
	if this.locked > 0 {
		atomic.StoreInt32(&this.locked, 0)
	}
}
func (this *MemoryDataset) Expire() int64 {
	return this.expire
}

func (this *MemoryDataset) indexOf(key string) int {
	for i, k := range this.keys {
		if k == key {
			return i
		}
	}
	return -1
}
func (this *MemoryDataset) append(key string, val interface{}) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if index := this.indexOf(key); index >= 0 {
		this.values[index] = val
	} else {
		this.keys = append(this.keys, key)
		this.values = append(this.values, val)
	}
}

func (this *MemoryDataset) GetArraySetKey() utils.ArraySetKey {
	v, _ := arraySetKeyDecode(this.id)
	return v
}

func (this *MemoryDataset) SetArraySetKey(arrayMapKey utils.ArraySetKey) {
	if this.id != "" {
		return //ID无法修改
	}
	this.id = arraySetKeyEncode(arrayMapKey)
}

func arraySetKeyEncode(arrayMapKey utils.ArraySetKey) string {
	return strconv.FormatInt(int64(arrayMapKey), memoryDatasetBitSize)
}

func arraySetKeyDecode(key string) (utils.ArraySetKey, error) {
	num, err := strconv.ParseInt(key, 10, memoryDatasetBitSize)
	if err != nil {
		return 0, err
	} else {
		return utils.ArraySetKey(num), nil
	}
}