package session

import (
	"github.com/hwcer/cosgo/utils"
	"time"
)

type Memory struct {
	stop chan struct{}
	*utils.ArraySet
}

func NewMemory() *Memory {
	return &Memory{
		ArraySet: utils.NewArraySet(int(Options.MapSize)),
	}
}

func (this *Memory) Start() {
	if Options.MaxAge > 0 {
		this.stop = make(chan struct{})
		go this.worker()
	}
}

func (this *Memory) Get(key string) (Dataset, bool) {
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
func (this *Memory) Create(data map[string]interface{}) Dataset {
	return NewMemoryDataset(data)
}

func (this *Memory) Remove(key string) bool {
	arrayMapKey, err := arraySetKeyDecode(key)
	if err != nil {
		return false
	}
	return this.ArraySet.Delete(arrayMapKey)
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
	this.ArraySet.Range(func(val utils.ArraySetVal) {
		if storage, ok := val.(*MemoryDataset); ok && storage.expire < nowTime {
			this.ArraySet.Delete(storage.GetArraySetKey())
		}
	})
}
