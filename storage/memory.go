package storage

import (
	"github.com/hwcer/cosgo/utils"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const memoryDatasetBitSize = 36

func NewMemory() *MemoryStorage {
	return &MemoryStorage{
		ArraySet: utils.NewArraySet(int(Options.MapSize)),
	}
}

func NewMemoryDataset(data map[string]interface{}) *MemoryDataset {
	d := &MemoryDataset{keys: make([]string, 0, len(data)), values: make([]interface{}, 0, len(data))}
	for k, v := range data {
		d.keys = append(d.keys, k)
		d.values = append(d.values, v)
	}
	d.Reset(false)
	return d
}

type MemoryStorage struct {
	stop chan struct{}
	*utils.ArraySet
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

func (this *MemoryDataset) Reset(locked bool) {
	if Options.MaxAge > 0 {
		this.expire = time.Now().Unix() + Options.MaxAge
	}
	if locked && this.locked > 0 {
		atomic.CompareAndSwapInt32(&this.locked, 1, 0)
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

func (this *MemoryStorage) Start() {
	if Options.MaxAge > 0 {
		this.stop = make(chan struct{})
		go this.worker()
	}
}

func (this *MemoryStorage) Get(key string) (Dataset, bool) {
	arrayMapKey, err := arraySetKeyDecode(key)
	if err != nil {
		return nil, false
	}
	val := this.ArraySet.Get(arrayMapKey)
	if val == nil {
		return nil, false
	}
	if s, ok := val.(*MemoryDataset); ok {
		return s, true
	} else {
		return nil, false
	}
}

//Create 创建新SESSION,返回SESSION ID
func (this *MemoryStorage) Create(data map[string]interface{}) Dataset {
	dataset := NewMemoryDataset(data)
	this.ArraySet.Add(dataset)
	return dataset
}

func (this *MemoryStorage) Remove(key string) bool {
	arrayMapKey, err := arraySetKeyDecode(key)
	if err != nil {
		return false
	}
	return this.ArraySet.Delete(arrayMapKey)
}
func (this *MemoryStorage) Close() {
	if Options.MaxAge == 0 || this.stop == nil {
		return
	}
	select {
	case <-this.stop:
	default:
		close(this.stop)
	}
}

func (this *MemoryStorage) worker() {
	ticker := time.NewTicker(time.Second * time.Duration(Options.Heartbeat))
	defer ticker.Stop()
	for {
		select {
		case <-this.stop:
			return
		case <-ticker.C:
			this.clean()
		}
	}
}

func (this *MemoryStorage) clean() {
	nowTime := time.Now().Unix()
	this.ArraySet.Range(func(val utils.ArraySetVal) {
		if storage, ok := val.(*MemoryDataset); ok && storage.expire < nowTime {
			this.ArraySet.Delete(storage.GetArraySetKey())
		}
	})
}
