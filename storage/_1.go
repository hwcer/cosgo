package storage

type Cache struct {
	cap       int
	bucket    []*Array
	mutex     sync.RWMutex //仅写操作需要锁
	prefix    string
	NewSetter func(id MID, val interface{}) Setter //创建新数据结构
}

// createSocketId 使用index生成ID
func (this *Cache) createId(index int) MID {
	this.seed++
	if this.seed >= math.MaxUint64 {
		this.seed = seedDefaultValue
		this.prefix = strconv.FormatInt(time.Now().Unix(), datasetKeyBitSize)
	}
	b := strings.Builder{}
	b.WriteString(strconv.FormatInt(int64(index), datasetKeyBitSize))
	b.WriteString(datasetKeySplit)
	b.WriteString(this.prefix)
	b.WriteString(datasetKeySplit)
	b.WriteString(strconv.FormatUint(this.seed, datasetKeyBitSize))
	return MID(b.String())
}

// parseSocketId 返回idPack中的index
func (this *Cache) parseId(id MID) int {
	s := string(id)
	index := strings.Index(s, datasetKeySplit)
	if index < 0 {
		return -1
	}
	i, err := strconv.ParseInt(s[0:index], datasetKeyBitSize, 32)
	if err != nil {
		return -1
	}
	return int(i)
}

func (this *Cache) get(id MID) (Setter, bool) {
	index := this.parseId(id)
	if index < 0 || index >= len(this.values) || this.values[index] == nil || this.values[index].Id() != id {
		return nil, false
	}
	return this.values[index], true
}

func (this *Cache) set(index int, val interface{}) (setter Setter) {
	size := len(this.values)
	if index < 0 || index > size {
		index = size
	}
	id := this.createId(index)
	setter = this.NewSetter(id, val)
	if index == size {
		this.values = append(this.values, setter) //扩容
	} else if this.values[index] == nil {
		this.values[index] = setter
	}
	return
}
func (this *Cache) push(v interface{}) Setter {
	if index := this.dirty.Acquire(); index >= 0 && index < len(this.values) && (this.values[index] == nil || this.values[index].Id() == "") {
		return this.set(index, v)
	}
	return this.set(-1, v)
}

func (this *Cache) remove(id MID) Setter {
	index := this.parseId(id)
	if index < 0 || index >= len(this.values) || this.values[index].Id() != id {
		return nil
	}
	val := this.values[index]
	this.values[index] = nil
	this.dirty.Release(index)
	return val
}

func (this *Cache) Get(id MID) (Setter, bool) {
	return this.get(id)
}

func (this *Cache) Set(id MID, v interface{}) bool {
	setter, ok := this.get(id)
	if !ok {
		return ok
	}
	setter.Set(v)
	return true
}

// Size 当前数量
func (this *Cache) Size() int {
	return len(this.values) - this.dirty.Size()
}

// Range 遍历
func (this *Cache) Range(f func(Setter) bool) {
	for _, val := range this.values {
		if val != nil && val.Id() != "" && !f(val) {
			break
		}
	}
}

// Mutex 获取操作锁
func (this *Cache) Mutex(f func()) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	f()
}

// Create 创建一个新数据
func (this *Cache) Create(v interface{}) Setter {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.push(v)
}

// Remove 批量移除
func (this *Cache) Remove(ids ...MID) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	for _, id := range ids {
		this.remove(id)
	}
	return
}

// Delete 删除并返回已删除的数据
func (this *Cache) Delete(id MID) Setter {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.remove(id)
}
