package session

import (
	"context"
	"github.com/hwcer/cosgo/storage"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const memoryDatasetBitSize = 36

func NewMemoryDataset(id uint64, data interface{}) storage.ArrayDataset {
	d := &MemoryDataset{}
	d.ArrayDatasetDefault = storage.NewArrayDataset(id, data).(*storage.ArrayDatasetDefault)
	if Options.MaxAge > 0 {
		d.Expire(Options.MaxAge)
	}
	return d
}

type MemoryDataset struct {
	lock   int32      //SESSION锁
	mutex  sync.Mutex //并发锁
	expire int64      //过期时间
	*storage.ArrayDatasetDefault
}

func (this *MemoryDataset) Lock() bool {
	return atomic.CompareAndSwapInt32(&this.lock, 0, 1)
}
func (this *MemoryDataset) UnLock() bool {
	return atomic.CompareAndSwapInt32(&this.lock, 1, 0)
}

//TTL 到期事件
func (this *MemoryDataset) TTL() int64 {
	return this.expire
}

//Expire 设置有效期(s)
func (this *MemoryDataset) Expire(s int64) {
	this.expire = time.Now().Unix() + s
}

func NewMemory() *Memory {
	s := &Memory{
		Array: storage.NewArray(1024),
	}
	s.Array.NewDataset = NewMemoryDataset
	return s
}

type Memory struct {
	stop chan struct{}
	*storage.Array
}

func (this *Memory) Start(ctx context.Context) error {
	if Options.MaxAge > 0 {
		this.stop = make(chan struct{})
		go this.worker(ctx)
	}
	return nil
}
func (this *Memory) get(key string) (dataset *MemoryDataset, err error) {
	var id uint64
	id, err = arrayKeyDecode(key)
	if err != nil {
		return nil, err
	}
	var ok bool
	var data storage.ArrayDataset
	if data, ok = this.Array.Dataset(id); !ok || data == nil {
		return nil, ErrorSessionNotExist
	}
	if dataset, ok = data.(*MemoryDataset); !ok {
		return nil, ErrorSessionTypeError
	}
	return dataset, nil
}

func (this *Memory) Get(sid string) (data map[string]interface{}, err error) {
	var ok bool
	var dataset *MemoryDataset
	if dataset, err = this.get(sid); err != nil {
		return nil, err
	}
	dataset.mutex.Lock()
	defer dataset.mutex.Unlock()
	var val map[string]interface{}
	if val, ok = dataset.Get().(map[string]interface{}); !ok {
		return nil, ErrorSessionTypeError
	}
	data = make(map[string]interface{}, len(val))

	for k, v := range val {
		data[k] = v
	}
	return data, nil
}

func (this *Memory) Set(sid string, data map[string]interface{}) (err error) {
	var ok bool
	var dataset *MemoryDataset
	if dataset, err = this.get(sid); err != nil {
		return err
	}
	dataset.mutex.Lock()
	defer dataset.mutex.Unlock()
	var val map[string]interface{}
	if val, ok = dataset.Get().(map[string]interface{}); !ok {
		return ErrorSessionTypeError
	}
	for k, v := range data {
		val[k] = v
	}
	return nil
}
func (this *Memory) Lock(sid string) bool {
	if dataset, err := this.get(sid); err == nil && dataset != nil {
		return dataset.Lock()
	} else {
		return false
	}
}

func (this *Memory) UnLock(sid string) bool {
	if dataset, err := this.get(sid); err == nil && dataset != nil {
		return dataset.UnLock()
	} else {
		return false
	}
}

//Create 创建新SESSION,返回SESSION ID
func (this *Memory) Create(data map[string]interface{}) string {
	id := this.Array.Set(data)
	return arrayKeyEncode(id)
}

func (this *Memory) Delete(sid string) bool {
	id, err := arrayKeyDecode(sid)
	if err != nil {
		return false
	}
	return this.Array.Delete(id)
}

//TTL 到期时间
func (this *Memory) TTL(sid string) (int64, error) {
	dataset, err := this.get(sid)
	if err != nil {
		return 0, err
	}
	return dataset.TTL(), nil
}

//Expire 设置有效期(s)
func (this *Memory) Expire(sid string) error {
	if Options.MaxAge == 0 {
		return nil
	}
	dataset, err := this.get(sid)
	if err != nil {
		return err
	}
	dataset.Expire(Options.MaxAge)
	return nil
}

func (this *Memory) Close() error {
	if Options.MaxAge == 0 || this.stop == nil {
		return nil
	}
	select {
	case <-this.stop:
	default:
		close(this.stop)
	}
	return nil
}

func (this *Memory) worker(ctx context.Context) {
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

func (this *Memory) clean() {
	nowTime := time.Now().Unix()
	this.Array.Range(func(item storage.ArrayDataset) bool {
		if storage, ok := item.(*MemoryDataset); ok && storage.expire < nowTime {
			this.Array.Delete(item.Id())
		}
		return true
	})
}

func arrayKeyEncode(arrayMapKey uint64) string {
	return strconv.FormatInt(int64(arrayMapKey), memoryDatasetBitSize)
}

func arrayKeyDecode(key string) (uint64, error) {
	num, err := strconv.ParseInt(key, memoryDatasetBitSize, 64)
	if err != nil {
		return 0, err
	} else {
		return uint64(num), nil
	}
}
