package session

import (
	"sync"
	"sync/atomic"
	"time"
)

type MemoryDataset struct {
	key      string
	mutex    sync.Mutex
	locked   int32
	expire   int64
	userdata map[string]interface{}
}

func (this *MemoryDataset) Set(val map[string]interface{}) {
	this.mutex.Lock()
	for k, v := range val {
		this.userdata[k] = v
	}
	this.mutex.Unlock()
}
func (this *MemoryDataset) Get() map[string]interface{} {
	val := make(map[string]interface{}, len(this.userdata))
	this.mutex.Lock()
	for k, v := range this.userdata {
		val[k] = v
	}
	this.mutex.Unlock()
	return val
}
func (this *MemoryDataset) Lock() bool {
	return atomic.CompareAndSwapInt32(&this.locked, 0, 1)
}
func (this *MemoryDataset) UnLock() {
	atomic.StoreInt32(&this.locked, 0)
}

type Memory struct {
	stop    chan struct{}
	data    map[string]*MemoryDataset
	lock    sync.Mutex
	options *Options
}

func NewMemory(opt *Options) *Memory {
	if opt == nil {
		opt = &Options{}
	}
	if opt.MapSize == 0 {
		opt.MapSize = 1000
	}
	if opt.Heartbeat == 0 {
		opt.Heartbeat = 10
	}
	m := &Memory{
		data:    make(map[string]*MemoryDataset, opt.MapSize),
		options: opt,
	}
	if opt.MaxAge > 0 {
		m.stop = make(chan struct{})
		go m.worker()
	}
	return m
}

func NewMemoryDataset(key string, data map[string]interface{}) *MemoryDataset {
	return &MemoryDataset{key: key, userdata: data}
}

func (this *Memory) Get(key string) (Dataset, bool) {
	this.lock.Lock()
	this.lock.Unlock()
	data, ok := this.data[key]
	if ok && this.options.MaxAge > 0 {
		nowTime := time.Now().Unix()
		if data.expire < nowTime {
			delete(this.data, key)
			return nil, false
		} else {
			data.expire = nowTime + this.options.MaxAge
		}
	}
	return data, ok
}

func (this *Memory) Set(key string, val map[string]interface{}) Dataset {
	this.lock.Lock()
	defer this.lock.Unlock()
	data, ok := this.data[key]
	if ok {
		data.Set(val)
	} else {
		data = NewMemoryDataset(key, val)
		data.locked = 1
		if this.options.MaxAge > 0 {
			data.expire = time.Now().Unix() + this.options.MaxAge
		}
		this.data[key] = data
	}
	return data
}

func (this *Memory) Del(key string) {
	this.lock.Lock()
	delete(this.data, key)
	this.lock.Unlock()
}
func (this *Memory) Close() {
	if this.options.MaxAge == 0 || this.stop == nil {
		return
	}
	select {
	case <-this.stop:
	default:
		close(this.stop)
	}
}

func (this *Memory) worker() {
	ticker := time.NewTicker(time.Second * time.Duration(this.options.Heartbeat))
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

func (this *Memory) clean() {
	var delKeys []string
	nowTime := time.Now().Unix()
	for _, v := range this.data {
		if v.expire <= nowTime {
			delKeys = append(delKeys, v.key)
		}
	}
	if len(delKeys) > 0 {
		this.lock.Lock()
		for _, k := range delKeys {
			delete(this.data, k)
		}
		this.lock.Unlock()
	}
}
