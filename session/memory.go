package session

import (
	"context"
	"github.com/hwcer/cosgo/scc"
	"github.com/hwcer/cosgo/session/storage"
	"github.com/hwcer/logger"
	"time"
)

func NewMemory(cap ...int) *Memory {
	var c int
	if len(cap) > 0 && cap[0] > 0 {
		c = cap[0]
	} else {
		c = 10240
	}
	s := &Memory{
		Storage: *storage.New(c, NewSetter),
	}
	//s.Array.NewSetter = NewSetter
	s.init()
	return s
}

type Memory struct {
	storage.Storage
}

func (this *Memory) init() {
	if Options.MaxAge > 0 {
		scc.CGO(this.daemon)
	}
}

func (this *Memory) get(id string) (*Setter, error) {
	if v, ok := this.Storage.Get(id); !ok {
		return nil, ErrorSessionIllegal
	} else {
		return v.(*Setter), nil
	}
}

func (this *Memory) New(p *Data) error {
	_ = this.Storage.Create(p)
	return nil
}

func (this *Memory) Verify(token string) (p *Data, err error) {
	var setter *Setter
	if setter, err = this.get(token); err == nil {
		p = setter.Data
		setter.KeepAlive()
	}
	return
}

// Update 更新信息，内存没事，共享Player信息已经更新过，仅仅设置过去时间
// 内存模式 data已经更新过，不需要再次更新

func (this *Memory) Update(p *Data, data map[string]any) (err error) {
	p.Update(data)
	return
}

func (this *Memory) Delete(d *Data) error {
	if setter := this.Storage.Delete(d.id); setter != nil {
		emit(d)
	}
	return nil
}

// Create 创建新SESSION,返回SESSION Index
// Create会自动设置有效期
// Create新数据为锁定状态
func (this *Memory) Create(uuid string, data map[string]any) (p *Data, err error) {
	p = NewData(uuid, data)
	_ = this.Storage.Create(p)
	return
}

func (this *Memory) daemon(ctx context.Context) {
	ts := time.Second * time.Duration(Options.Heartbeat)
	ticker := time.NewTimer(ts)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			this.clean()
			ticker.Reset(ts)
		}
	}
}

func (this *Memory) clean() {
	defer func() {
		if err := recover(); err != nil {
			logger.Alert("session.memory daemon ticker error:%v", err)
		}
	}()
	var keys []string
	maxAge := int32(Options.MaxAge)
	this.Storage.Range(func(item storage.Setter) bool {
		if data, ok := item.(*Setter); ok && data.Heartbeat(Options.Heartbeat) >= maxAge {
			keys = append(keys, data.Id())
		}
		return true
	})

	for _, key := range keys {
		if setter := this.Storage.Delete(key); setter != nil {
			if v := setter.(*Setter); v != nil {
				emit(v.Data)
			}
		}
	}
}
