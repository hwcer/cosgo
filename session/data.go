package session

import (
	"math"
	"sync"
	"sync/atomic"

	"github.com/hwcer/cosgo/values"
)

func NewData(uuid string, vs map[string]any) *Data {
	p := &Data{id: uuid, uuid: uuid}
	if len(vs) > 0 {
		p.Values = vs
	} else {
		p.Values = values.Values{}
	}
	return p
}

var MaxDataIndex = int32(math.MaxInt32 - 1000)

// Data 用户登录信息,不要直接修改 Player.Values 信息
type Data struct {
	id    string // 默认uuid,memory模式会定制此ID
	uuid  string //GUID
	index int32  //socket server id
	values.Values
	heartbeat int32
	mutex     sync.Mutex
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

func (this *Data) set(key string, value any) any {
	vs := this.Values.Clone()
	vs.Set(key, value)
	this.Values = vs
	return value
}

// update 批量设置Cookie信息
func (this *Data) update(data map[string]any) {
	vs := this.Values.Clone()
	for k, v := range data {
		vs.Set(k, v)
	}
	this.Values = vs
}
func (this *Data) delete(key string) {
	vs := this.Values.Clone()
	delete(vs, key)
	this.Values = vs
}

func (this *Data) Set(key string, value any) any {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.set(key, value)
}

// Update 批量设置Cookie信息
func (this *Data) Update(data map[string]any) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.update(data)
}
func (this *Data) Delete(key string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.delete(key)
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

// Mutex 获得写权限，注意必须使用Setter进行操作
func (this *Data) Mutex(cb func(Setter)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	s := Setter{Data: this}
	cb(s)
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
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if this.index >= MaxDataIndex {
		this.index = 0
	}
}

type Setter struct {
	*Data
}

func (this *Setter) Set(key string, value any) any {
	return this.set(key, value)
}
func (this *Setter) Update(data map[string]any) {
	this.update(data)
}
func (this *Setter) Delete(key string) {
	this.delete(key)
}
