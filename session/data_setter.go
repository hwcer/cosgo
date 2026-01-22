package session

// Mutex 获得写权限，注意必须使用Setter进行操作
func (this *Data) Mutex(cb func(Setter)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	s := Setter{Data: this}
	cb(s)
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
