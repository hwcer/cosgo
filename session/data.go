package session

import (
	"github.com/hwcer/cosgo/values"
	"math"
	"sync"
	"sync/atomic"
)

func NewData(uuid string, vs map[string]any, token ...string) *Data {
	p := &Data{uuid: uuid}
	if len(vs) > 0 {
		p.Values = vs
	} else {
		p.Values = values.Values{}
	}
	if len(token) > 0 {
		p.id = token[0]
	}
	return p
}

var MaxDataIndex = int32(math.MaxInt32 - 1000)

// Data 用户登录信息,不要直接修改 Player.Values 信息
type Data struct {
	id    string //token
	uuid  string //GUID
	index int32  //socket server id
	sync.Mutex
	values.Values
	heartbeat int32
}

func (this *Data) Id() string {
	if this == nil {
		return ""
	}
	return this.id
}

func (this *Data) Is(v *Data) bool {
	if v == nil {
		return false
	}
	return this.id == v.id
}
func (this *Data) KeepAlive() {
	this.heartbeat = 0
}
func (this *Data) Heartbeat(v ...int32) int32 {
	if len(v) > 0 {
		this.heartbeat += v[0]
	}
	return this.heartbeat
}

func (this *Data) Set(key string, value any, locked ...bool) any {
	if !(len(locked) > 0 && locked[0]) {
		this.Lock()
		defer this.Unlock()
	}

	vs := this.Values.Clone()
	vs.Set(key, value)
	this.Values = vs
	return value
}

func (this *Data) Merge(p *Data, locked ...bool) {
	if this.id == p.id {
		return
	}
	if !(len(locked) > 0 && locked[0]) {
		this.Lock()
		defer this.Unlock()
	}
	vs := this.Values.Clone()
	vs.Merge(p.Values, false)
	this.Values = vs
}

// Update 批量设置Cookie信息
func (this *Data) Update(data map[string]any, locked ...bool) {
	if !(len(locked) > 0 && locked[0]) {
		this.Lock()
		defer this.Unlock()
	}
	vs := this.Values.Clone()
	for k, v := range data {
		vs.Set(k, v)
	}
	this.Values = vs
}

func (this *Data) UUID() string {
	if this == nil {
		return ""
	}
	return this.uuid
}

func (this *Data) Reset() {
	this.index = 0
}

// Atomic 生成一个自增的包序列号
func (this *Data) Atomic() int32 {
	if this.index >= MaxDataIndex {
		this.TryResetIndex()
	}
	r := atomic.AddInt32(&this.index, 1)
	return r
}
func (this *Data) Index() int32 {
	return this.index
}

func (this *Data) TryResetIndex() {
	this.Lock()
	defer this.Unlock()
	if this.index >= MaxDataIndex {
		this.index = 0
	}
}
