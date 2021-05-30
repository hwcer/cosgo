package session

import (
	"github.com/hwcer/cosgo/utils"
	"time"
)

type Memory struct {
	stop    chan struct{}
	options *Options
	*utils.ArrayMap
}

func NewMemory(opt *Options) *Memory {
	if opt == nil {
		opt = &Options{}
	}
	if opt.MapSize == 0 {
		opt.MapSize = 1024
	}
	if opt.Heartbeat == 0 {
		opt.Heartbeat = 10
	}
	m := &Memory{
		options:  opt,
		ArrayMap: utils.NewArrayMap(int(opt.MapSize)),
	}
	if opt.MaxAge > 0 {
		m.stop = make(chan struct{})
		go m.worker()
	}
	return m
}

func (this *Memory) Get(key string) (*Storage, bool) {
	arrayMapKey, err := ArrayMapKeyDecode(key)
	if err != nil {
		return nil, false
	}
	val := this.ArrayMap.Get(arrayMapKey)
	if val == nil {
		return nil, false
	}
	if s, ok := val.(*Storage); ok {
		return s, true
	} else {
		return nil, false
	}
}

//Set 设置修改SESSION的内容
func (this *Memory) Set(key string, data map[string]interface{}) bool {
	storage, ok := this.Get(key)
	if !ok {
		return false
	}
	if this.options.MaxAge > 0 {
		data[StorageExpireKey] = time.Now().Unix() + this.options.MaxAge
	}
	storage.Set(data)
	return true
}

//New 创建新SESSION,返回SESSION ID
func (this *Memory) Ceate(data map[string]interface{}) *Storage {
	storage := NewStorage(data)
	if this.options.MaxAge > 0 {
		data[StorageExpireKey] = time.Now().Unix() + this.options.MaxAge
	}
	arrayMapKey := this.ArrayMap.Add(storage)
	storage.SetArrayMapKey(arrayMapKey)
	return storage
}

func (this *Memory) Remove(key string) bool {
	arrayMapKey, err := ArrayMapKeyDecode(key)
	if err != nil {
		return false
	}
	return this.ArrayMap.Remove(arrayMapKey)
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
	nowTime := time.Now().Unix()
	this.ArrayMap.Range(func(val utils.ArrayMapVal) {
		if storage, ok := val.(*Storage); ok && storage.expire < nowTime {
			this.ArrayMap.Remove(storage.GetArrayMapKey())
		}
	})
}
