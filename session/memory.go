// Package session 提供会话管理功能，支持内存和Redis存储
package session

import (
	"github.com/hwcer/cosgo/storage"
	"github.com/hwcer/logger"
)

// 注意：
// 1. 内存存储实现中，会话数据存储在内存中，适用于单机应用
// 2. 内置心跳机制，自动清理过期会话
// 3. 内存模式下，Data 已经在写操作时更新过，Update 方法不需要再次更新

func NewMemory(cap ...int) *Memory {
	var c int
	if len(cap) > 0 && cap[0] > 0 {
		c = cap[0]
	} else {
		c = 10240
	}
	s := &Memory{
		Storage: *storage.New(c, NewMemorySetter),
	}
	s.init()
	return s
}

type Memory struct {
	storage.Storage
}

func (this *Memory) init() {
	if Options.MaxAge > 0 && Options.Heartbeat > 0 {
		On(EventHeartbeat, this.Heartbeat)
		Heartbeat.Start()
	}
}

func (this *Memory) get(id string) (*Data, error) {
	v, ok := this.Storage.Get(id)
	if !ok {
		return nil, ErrorSessionIllegal
	}
	return v.(*Data), nil
}

func (this *Memory) New(p *Data) error {
	_ = this.Storage.New(p)
	return nil
}

func (this *Memory) Get(id string) (data *Data, err error) {
	if data, err = this.get(id); err == nil {
		data.KeepAlive()
	}
	return
}

// Update 更新信息，内存没事，共享Player信息已经更新过，仅仅设置过期时间
// 内存模式 data已经更新过，不需要再次更新
func (this *Memory) Update(p *Data, data map[string]any) (err error) {
	//p.Update(data)
	return
}

func (this *Memory) Delete(d *Data) error {
	_ = this.Storage.Delete(d.id)
	return nil
}

// Create 创建新SESSION
func (this *Memory) Create(uuid string, data map[string]any) (p *Data, err error) {
	p = NewData(uuid, data)
	_ = this.Storage.New(p)
	return
}

func (this *Memory) Heartbeat(i any) {
	s, _ := i.(int32)
	if s <= 0 {
		return
	}
	defer func() {
		if err := recover(); err != nil {
			logger.Alert("session.memory daemon ticker error:%v", err)
		}
	}()
	var vs []*Data
	maxAge := int32(Options.MaxAge)
	this.Storage.Range(func(item storage.Setter) bool {
		if d, ok := item.(*Data); ok && d.Heartbeat(s) >= maxAge {
			vs = append(vs, d)
		}
		return true
	})

	for _, v := range vs {
		_ = this.Delete(v)
		//内部删除想要触发事件
		Emit(EventSessionRelease, v)
	}
}
