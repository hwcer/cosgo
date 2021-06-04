package storage

import (
	"context"
	"github.com/hwcer/cosgo/cosmap"
	"strconv"
	"sync/atomic"
	"time"
)

const memoryDatasetBitSize = 36

func NewMemory() *MemoryStorage {
	return &MemoryStorage{
		Array: cosmap.NewArray(int(Options.MapSize)),
	}
}

func NewMemoryDataset(data map[string]interface{}) *MemoryDataset {
	d := &MemoryDataset{SRMap: *cosmap.NewSRMap(len(data))}
	d.SRMap.MSet(data)
	d.Reset(false)
	return d
}

type MemoryStorage struct {
	stop chan struct{}
	*cosmap.Array
}

type MemoryDataset struct {
	id     string
	locked int32
	expire int64
	cosmap.SRMap
}

//Id 获取session id
func (this *MemoryDataset) Id() string {
	return this.id
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

func (this *MemoryDataset) GetArrayKey() cosmap.ArrayKey {
	v, _ := arraySetKeyDecode(this.id)
	return v
}

func (this *MemoryDataset) SetArrayKey(arrayMapKey cosmap.ArrayKey) {
	if this.id != "" {
		return //ID无法修改
	}
	id := arraySetKeyEncode(arrayMapKey)
	this.id = id
}

func arraySetKeyEncode(arrayMapKey cosmap.ArrayKey) string {
	return strconv.FormatInt(int64(arrayMapKey), memoryDatasetBitSize)
}

func arraySetKeyDecode(key string) (cosmap.ArrayKey, error) {
	num, err := strconv.ParseInt(key, memoryDatasetBitSize, 64)
	if err != nil {
		return 0, err
	} else {
		return cosmap.ArrayKey(num), nil
	}
}

func (this *MemoryStorage) Start(ctx context.Context) {
	if Options.MaxAge > 0 {
		this.stop = make(chan struct{})
		go this.worker(ctx)
	}
}

func (this *MemoryStorage) Get(key string) (Dataset, bool) {
	arrayMapKey, err := arraySetKeyDecode(key)
	if err != nil {
		return nil, false
	}
	val := this.Array.Get(arrayMapKey)
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
	this.Array.Add(dataset)
	return dataset
}

func (this *MemoryStorage) Delete(key string) bool {
	arrayMapKey, err := arraySetKeyDecode(key)
	if err != nil {
		return false
	}
	return this.Array.Delete(arrayMapKey)
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

func (this *MemoryStorage) worker(ctx context.Context) {
	ticker := time.NewTicker(time.Second * time.Duration(Options.Heartbeat))
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-this.stop:
			return
		case <-ticker.C:
			this.clean()
		}
	}
}

func (this *MemoryStorage) clean() {
	nowTime := time.Now().Unix()
	this.Array.Range(func(val cosmap.ArrayVal) {
		if storage, ok := val.(*MemoryDataset); ok && storage.expire < nowTime {
			this.Array.Delete(storage.GetArrayKey())
		}
	})
}
