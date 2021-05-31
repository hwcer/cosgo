package session

import (
	"github.com/hwcer/cosgo/utils"
	"time"
)

type Memory struct {
	stop chan struct{}
	*utils.ArrayMap
}

func NewMemory() *Memory {
	return &Memory{
		ArrayMap: utils.NewArrayMap(int(Options.MapSize)),
	}
}

func (this *Memory) Start() {
	if Options.MaxAge > 0 {
		this.stop = make(chan struct{})
		go this.worker()
	}
}

func (this *Memory) Get(key string) (Dataset, bool) {
	arrayMapKey, err := arrayMapKeyDecode(key)
	if err != nil {
		return nil, false
	}
	val := this.ArrayMap.Get(arrayMapKey)
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
func (this *Memory) Create(data map[string]interface{}) Dataset {
	storage := NewMemoryDataset(data)
	arrayMapKey := this.ArrayMap.Add(storage)
	storage.SetArrayMapKey(arrayMapKey)
	return storage
}

func (this *Memory) Remove(key string) bool {
	arrayMapKey, err := arrayMapKeyDecode(key)
	if err != nil {
		return false
	}
	return this.ArrayMap.Remove(arrayMapKey)
}
func (this *Memory) Close() {
	if Options.MaxAge == 0 || this.stop == nil {
		return
	}
	select {
	case <-this.stop:
	default:
		close(this.stop)
	}
}

func (this *Memory) worker() {
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

func (this *Memory) clean() {
	nowTime := time.Now().Unix()
	this.ArrayMap.Range(func(val utils.ArrayMapVal) {
		if storage, ok := val.(*MemoryDataset); ok && storage.expire < nowTime {
			this.ArrayMap.Remove(storage.GetArrayMapKey())
		}
	})
}
