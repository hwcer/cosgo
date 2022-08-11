package smap

import "sync"

//基于Array的键值对
//仅仅 Create Delete 是使用固定的key(uid) 防止重复登录
//Get Set 仍然希望使用MID达到无锁状态

func NewHash(cap int) *Hash {
	return &Hash{Array: New(cap), keys: sync.Map{}}
}

type Hash struct {
	*Array
	keys sync.Map
}

func (this *Hash) Load(key string) MID {
	if mid, ok := this.keys.Load(key); ok {
		return mid.(MID)
	}
	return ""
}

func (this *Hash) Create(key string, val interface{}) Setter {
	if mid, ok := this.keys.Load(key); ok {
		this.Array.Delete(mid.(MID))
	}
	setter := this.Array.Push(val)
	this.keys.Store(key, setter.Id())
	return setter
}

func (this *Hash) Delete(key string) (r Setter) {
	if mid, loaded := this.keys.LoadAndDelete(key); loaded {
		r = this.Array.Delete(mid.(MID))
	}
	return
}

func (this *Hash) Remove(keys ...string) {
	for _, k := range keys {
		this.Delete(k)
	}
}
