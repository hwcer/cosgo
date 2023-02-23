package storage

import "sync"

//基于Array的键值对
//仅仅 Create delete 时使用固定的key(uid) 防止重复登录
//Value Set 仍然希望使用MID达到无锁状态

func NewHash(cap int) *Hash {
	return &Hash{Array: *New(cap), keys: sync.Map{}}
}

type Hash struct {
	Array
	keys sync.Map
}

func (this *Hash) MID(uuid string) (mid MID, ok bool) {
	var v interface{}
	if v, ok = this.keys.Load(uuid); !ok {
		return
	}
	mid, ok = v.(MID)
	return
}

// Reset 重新设置uuid mid关联，返回之前的数据(如果存在)
func (this *Hash) Reset(uuid string, mid MID) (r Setter) {
	v, ok := this.MID(uuid)
	if ok && v == mid {
		return
	}
	this.keys.Store(uuid, mid)
	if ok {
		r = this.Array.remove(v)
	}
	return
}

func (this *Hash) Load(uuid string) (r Setter) {
	if mid, ok := this.MID(uuid); ok {
		r, _ = this.Array.Get(mid)
	}
	return
}

func (this *Hash) Create(uuid string, val interface{}) Setter {
	setter := this.Array.Create(val)
	_ = this.Reset(uuid, setter.Id())
	return setter
}

func (this *Hash) Delete(uuid string) (r Setter) {
	v, loader := this.keys.LoadAndDelete(uuid)
	if !loader {
		return
	}
	if mid, ok := v.(MID); ok {
		r = this.Array.remove(mid)
	}
	return
}
func (this *Hash) Remove(uuid ...string) {
	var mid []MID
	for _, uid := range uuid {
		m, loader := this.keys.LoadAndDelete(uid)
		if loader {
			if v, ok := m.(MID); ok {
				mid = append(mid, v)
			}
		}
	}
	if len(mid) > 0 {
		this.Array.Remove(mid...)
	}
	return
}
