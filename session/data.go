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
	id        string        // 默认uuid,memory模式会定制此ID
	uuid      string        //GUID
	index     int32         //socket server id
	values    values.Values // 私有字段，外部通过方法访问
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

func (this *Data) Set(key string, value any) any {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.set(key, value)
}

func (this *Data) Values() values.Values {
	return this.values.Clone()
}

// Get 获取指定键的值
func (this *Data) Get(key string) any {
	return this.values.Get(key)
}

// GetString 获取指定键的字符串值
func (this *Data) GetString(key string) string {
	return this.values.GetString(key)
}

// GetInt 获取指定键的整数值
func (this *Data) GetInt(key string) int {
	return this.values.GetInt(key)
}
func (this *Data) GetInt32(key string) int32 {
	return this.values.GetInt32(key)
}

// GetInt64 获取指定键的64位整数值
func (this *Data) GetInt64(key string) int64 {
	return this.values.GetInt64(key)
}

// GetFloat64 获取指定键的浮点数值
func (this *Data) GetFloat64(key string) float64 {
	return this.values.GetFloat64(key)
}

// Range 遍历所有键值对
func (this *Data) Range(cb func(key string, value any) bool) {
	this.values.Range(cb)
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
