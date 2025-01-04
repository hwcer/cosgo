package session

import (
	"context"
	"github.com/hwcer/cosgo/logger"
	"github.com/hwcer/cosgo/scc"
	"github.com/hwcer/cosgo/storage"
	"time"
)

var Heartbeat int32 = 10 //心跳(S)

func NewMemory() *Memory {
	s := &Memory{
		Array: *storage.New(1024),
	}
	s.Array.NewSetter = NewSetter
	s.init()
	return s
}

type Memory struct {
	storage.Array
}

func (this *Memory) init() {
	if Options.MaxAge > 0 {
		scc.CGO(this.daemon)
	}
}

func (this *Memory) get(token string) (*Setter, error) {
	mid := storage.MID(token)
	if v, ok := this.Array.Get(mid); !ok {
		return nil, ErrorSessionIllegal
	} else {
		return v.(*Setter), nil
	}
}

func (this *Memory) Verify(token string) (p *Data, err error) {
	var setter *Setter
	if setter, err = this.get(token); err == nil {
		p, _ = setter.Get().(*Data)
	}
	return
}

// Update 更新信息，内存没事，共享Player信息已经更新过，仅仅设置过去时间
// 内存模式 data已经更新过，不需要再次更新

func (this *Memory) Update(p *Data, data map[string]any, ttl int64) (err error) {
	var setter *Setter
	setter, err = this.get(p.token)
	if err != nil {
		return
	}

	if ttl > 0 {
		setter.Expire(ttl)
	}
	return
}
func (this *Memory) Delete(p *Data) error {
	mid := storage.MID(p.token)
	this.Array.Delete(mid)
	return nil
}

// Create 创建新SESSION,返回SESSION Index
// Create会自动设置有效期
// Create新数据为锁定状态
func (this *Memory) Create(uuid string, data map[string]any, ttl int64) (p *Data, err error) {
	d := this.Array.Create(nil)
	setter, _ := d.(*Setter)
	st := string(setter.Id())
	p = NewData(uuid, st, data)
	setter.Set(p)
	if ttl > 0 {
		setter.Expire(ttl)
	}
	return
}

func (this *Memory) daemon(ctx context.Context) {
	ticker := time.NewTimer(time.Second * time.Duration(Heartbeat))
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			this.clean()
		}
	}
}

func (this *Memory) clean() {
	defer func() {
		if err := recover(); err != nil {
			logger.Alert("session.memory daemon ticker error:%v", err)
		}
	}()
	nowTime := time.Now().Unix()
	var keys []storage.MID
	this.Array.Range(func(item storage.Setter) bool {
		if data, ok := item.(*Setter); ok && data.expire < nowTime {
			keys = append(keys, data.Id())
		}
		return true
	})
	if len(keys) > 0 {
		this.Remove(keys...)
	}
}
