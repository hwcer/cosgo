package session

import (
	"github.com/hwcer/cosgo/session/storage"
	"github.com/hwcer/logger"
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
	listeners []func(data *Data)
}

func (this *Memory) init() {
	if Options.MaxAge > 0 && Options.Heartbeat > 0 {
		Heartbeat.On(this.Heartbeat)
		Heartbeat.Start()
	}
}

func (this *Memory) get(id string) (*Setter, error) {
	if v, ok := this.Storage.Get(id); !ok {
		return nil, ErrorSessionIllegal
	} else {
		return v.(*Setter), nil
	}
}

func (this *Memory) emit(v *Data) {
	for _, l := range this.listeners {
		l(v)
	}
}

func (this *Memory) On(l func(data *Data)) {
	this.listeners = append(this.listeners, l)
}

func (this *Memory) New(p *Data) error {
	_ = this.Storage.Create(p)
	return nil
}

func (this *Memory) Get(id string) (p *Data, err error) {
	var setter *Setter
	if setter, err = this.get(id); err == nil {
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
		this.emit(d)
	}
	return nil
}

// Create 创建新SESSION,返回SESSION Index
// Create会自动设置有效期
// Create新数据为锁定状态
func (this *Memory) Create(uuid string, data map[string]any) (p *Data, err error) {
	p = NewData(uuid, data) //setter中自动设置
	_ = this.Storage.Create(p)
	return
}

func (this *Memory) Heartbeat(s int32) {
	defer func() {
		if err := recover(); err != nil {
			logger.Alert("session.memory daemon ticker error:%v", err)
		}
	}()
	var vs []*Data
	maxAge := int32(Options.MaxAge)
	this.Storage.Range(func(item storage.Setter) bool {
		if d, ok := item.(*Setter); ok && d.Heartbeat(s) >= maxAge {
			vs = append(vs, d.Data)
		}
		return true
	})

	for _, v := range vs {
		_ = this.Delete(v)
	}
}
