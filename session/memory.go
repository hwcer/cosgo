package session

import (
	"context"
	"github.com/hwcer/cosgo/scc"
	"github.com/hwcer/cosgo/session/storage"
	"github.com/hwcer/logger"
	"time"
)

var Heartbeat int32 = 10 //心跳(S)

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
func (this *Memory) Delete(p *Data) error {
	this.Storage.Delete(p.id)
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
	var keys []string
	this.Storage.Range(func(item storage.Setter) bool {
		if data, ok := item.(*Setter); ok && data.expire < nowTime {
			keys = append(keys, data.Id())
		}
		return true
	})

	if len(keys) > 0 {
		this.Remove(keys)
	}
}
