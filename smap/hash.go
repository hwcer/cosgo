package smap

//基于Array的键值对
//仅仅 Create delete 是使用固定的key(uid) 防止重复登录
//Value Set 仍然希望使用MID达到无锁状态

func NewHash(cap int) *Hash {
	return &Hash{Array: New(cap), keys: make(map[string]MID)}
}

type Hash struct {
	*Array
	keys map[string]MID
}

func (this *Hash) MID(uuid string) (MID, bool) {
	this.Array.mutex.RLock()
	defer this.Array.mutex.RUnlock()
	mid, ok := this.keys[uuid]
	return mid, ok
}

// Reset 重新设置uuid mid关联，返回之前的数据(如果存在)
func (this *Hash) Reset(uuid string, mid MID) (r Setter) {
	this.Array.mutex.RLock()
	defer this.Array.mutex.RUnlock()
	if k, ok := this.keys[uuid]; ok {
		r, _ = this.Array.Get(k)
	}
	this.keys[uuid] = mid
	return
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
