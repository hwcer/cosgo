// Package session 提供会话管理功能，支持内存和Redis存储
package session

import (
	"math"
	"sync"
	"sync/atomic"

	"github.com/hwcer/cosgo/values"
)

// 注意：
// 1. Data 结构中的 values 字段是私有的，外部只能通过提供的方法访问
// 2. 读操作（如 Get、GetString 等）是无锁的，直接访问 values 字段
// 3. 写操作（如 Set、Update、Delete 等）使用互斥锁保护，并采用 Copy-on-Write 模式
// 4. 一个 Session 绑定的是一个用户的单次请求的上下文，不会存在并发问题

func NewData(uuid string, vs map[string]any) *Data {
	p := &Data{id: uuid, uuid: uuid}
	if len(vs) > 0 {
		p.values = vs
	} else {
		p.values = values.Values{}
	}
	return p
}

var MaxDataIndex = int32(math.MaxInt32 - 1000)

// Data 用户登录信息,不要直接修改 Player.Values 信息
type Data struct {
	id        string        // session id
	uuid      string        // GUID
	index     int32         // socket server id
	values    values.Values //私有字段，外部通过方法访问
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
	vs := this.values.Clone()
	vs.Set(key, value)
	this.values = vs
	return value
}

// update 批量设置Cookie信息
func (this *Data) update(data map[string]any) {
	vs := this.values.Clone()
	for k, v := range data {
		vs.Set(k, v)
	}
	this.values = vs
}
func (this *Data) delete(key string) {
	vs := this.values.Clone()
	delete(vs, key)
	this.values = vs
}

// Set 设置Cookie信息
// done 可选参数，用于在设置完成后锁内安全执行额外操作
func (this *Data) Set(key string, value any, done ...func()) any {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	r := this.set(key, value)
	if len(done) > 0 {
		done[0]()
	}
	return r
}

func (this *Data) Values() values.Values {
	return this.values.Clone()
}

// Range 遍历所有键值对
func (this *Data) Range(cb func(key string, value any) bool) {
	this.values.Range(cb)
}

// Update 批量设置Cookie信息
// done 可选参数，用于在设置完成后锁内安全执行额外操作
func (this *Data) Update(data map[string]any, done ...func()) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.update(data)
	if len(done) > 0 {
		done[0]()
	}
}
func (this *Data) Delete(key string, done ...func()) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.delete(key)
	if len(done) > 0 {
		done[0]()
	}
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
		this.tryResetIndex()
	}
	r := atomic.AddInt32(&this.index, 1)
	return r
}
func (this *Data) Index() int32 {
	return this.index
}

func (this *Data) tryResetIndex() {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if this.index >= MaxDataIndex {
		this.index = 0
	}
}
