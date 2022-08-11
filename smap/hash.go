package smap

//基于Array的键值对
//仅仅 Create Delete 是使用固定的key(uid) 防止重复登录
//Get Set 仍然希望使用MID达到无锁状态

func NewHash(cap int) *Hash {
	return &Hash{Array: *New(cap), keys: make(map[string]MID)}
}

type Hash struct {
	Array
	keys map[string]MID
}

func (this *Hash) MID(uuid string) (MID, bool) {
	this.Array.mutex.Lock()
	defer this.Array.mutex.Unlock()
	mid, ok := this.keys[uuid]
	return mid, ok
}

func (this *Hash) Create(uuid string, val interface{}) Setter {
	this.Array.mutex.Lock()
	defer this.Array.mutex.Unlock()
	if mid, ok := this.keys[uuid]; ok {
		this.Array.remove(mid)
	}
	setter := this.Array.push(val)
	this.keys[uuid] = setter.Id()
	return setter
}

func (this *Hash) Delete(uuid string) (r Setter) {
	this.Array.mutex.Lock()
	defer this.Array.mutex.Unlock()
	if mid, ok := this.keys[uuid]; ok {
		this.Array.remove(mid)
	}
	delete(this.keys, uuid)
	return
}

func (this *Hash) Remove(uuid ...string) {
	this.Array.mutex.Lock()
	defer this.Array.mutex.Unlock()
	for _, key := range uuid {
		if mid, ok := this.keys[key]; ok {
			this.Array.remove(mid)
		}
		delete(this.keys, key)
	}
}
