package memory

import (
	"github.com/hwcer/cosgo/session/options"
	"github.com/hwcer/cosgo/storage/cache"
	"time"
)

var Options = &options.Options //导入配置
var Heartbeat int32 = 10       //心跳(S)

func New() *Memory {
	s := &Memory{
		Cache: cache.New(1024),
	}
	s.Cache.NewSetter = NewUpdater
	return s
}

type Memory struct {
	*cache.Cache
	stop chan struct{}
}

func (this *Memory) Start() error {
	if Options.MaxAge > 0 {
		this.stop = make(chan struct{})
		go this.worker()
	}
	return nil
}
func (this *Memory) get(key string) (*Dataset, error) {
	var (
		ok  bool
		id  uint64
		err error
	)
	if id, err = cache.Decode(key); err != nil {
		return nil, err
	}
	var data cache.Dataset
	if data, ok = this.Cache.Get(id); !ok || data == nil {
		return nil, options.ErrorSessionNotExist
	}
	var val *Dataset
	if val, ok = data.(*Dataset); !ok {
		return nil, options.ErrorSessionTypeError
	}
	if val.expire > 0 && val.expire < time.Now().Unix() {
		return nil, options.ErrorSessionTypeExpire
	}
	return val, nil
}

func (this *Memory) Get(key string, lock bool) (uid string, result map[string]interface{}, err error) {
	var ok bool
	var data *Dataset
	if data, err = this.get(key); err != nil {
		return "", nil, err
	}
	if lock && !data.Lock() {
		return "", nil, options.ErrorSessionLocked
	}
	uid = data.uid
	var val map[string]interface{}
	if val, ok = data.Get().(map[string]interface{}); !ok {
		return "", nil, options.ErrorSessionTypeError
	}
	result = make(map[string]interface{}, len(val))
	for k, v := range val {
		result[k] = v
	}
	return
}

func (this *Memory) Save(key string, data map[string]interface{}, expire int64, unlock bool) (err error) {
	var setter *Dataset
	if setter, err = this.get(key); err != nil {
		return err
	}
	var ok bool
	var value map[string]interface{}
	if value, ok = setter.Get().(map[string]interface{}); !ok {
		return options.ErrorSessionTypeError
	}
	result := make(map[string]interface{})
	for k, v := range value {
		result[k] = v
	}
	for k, v := range data {
		result[k] = v
	}
	setter.Set(result)
	if expire > 0 {
		setter.Expire(expire)
	}
	if unlock {
		setter.UnLock()
	}
	return
}

//Create 创建新SESSION,返回SESSION Index
//Create会自动设置有效期
//Create新数据为锁定状态
func (this *Memory) Create(uid string, data map[string]interface{}, expire int64, lock bool) (sid, key string, err error) {
	id := this.Cache.Push(data)
	key = cache.Encode(id)
	sid, err = options.Encode(key)
	var setter *Dataset
	if setter, err = this.get(key); err != nil {
		return "", "", err
	}
	setter.uid = uid
	if expire > 0 {
		setter.Expire(expire)
	}
	if lock {
		setter.Lock()
	}
	return
}

func (this *Memory) Delete(key string) error {
	id, err := cache.Decode(key)
	if err != nil {
		return err
	}
	this.Cache.Delete(id)
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

func (this *Memory) worker() {
	ticker := time.NewTicker(time.Second * time.Duration(Heartbeat))
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
	var remove []uint64
	this.Cache.Range(func(item cache.Dataset) bool {
		if data, ok := item.(*Dataset); ok && data.expire < nowTime {
			remove = append(remove, item.Id())
		}
		return true
	})
	if len(remove) > 0 {
		this.Cache.Remove(remove...)
	}
}
