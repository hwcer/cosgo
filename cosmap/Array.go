package cosmap

import "sync"

type ArrayKey uint64
type ArrayVal interface {
	GetArrayKey() ArrayKey //获取ArraySet Key
	SetArrayKey(ArrayKey)  //设置ArraySet Key
}

func NewArray(cap int) *Array {
	arrayMap := &Array{
		seed:   1,
		remove: NewArrayIndex(cap),
		values: make([]ArrayVal, cap, cap),
	}
	for i := cap - 1; i >= 0; i-- {
		arrayMap.remove.Add(i)
	}
	return arrayMap
}

type Array struct {
	seed   uint32 //ID 生成种子
	mutex  sync.Mutex
	values []ArrayVal
	remove *ArrayIndex
}

//createSocketId 使用index生成ID
func (s *Array) createId(index int) ArrayKey {
	s.seed++
	return ArrayKey(index)<<32 | ArrayKey(s.seed)
}

//parseSocketId 返回idPack中的index
func (s *Array) parseId(id ArrayKey) int {
	return int(id >> 32)
}

func (s *Array) Add(v ArrayVal) ArrayKey {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	var index = -1
	if index = s.remove.Get(); index >= 0 {
		s.values[index] = v
	} else {
		index = len(s.values)
		s.values = append(s.values, v)
	}
	id := s.createId(index)
	v.SetArrayKey(id)
	return id
}

//Get 获取
func (s *Array) Get(id ArrayKey) ArrayVal {
	index := s.parseId(id)
	if index >= len(s.values) {
		return nil
	}
	if val := s.values[index]; val != nil && val.GetArrayKey() == id {
		return val
	} else {
		return nil
	}
}

//Delete 删除
func (s *Array) Delete(id ArrayKey) bool {
	index := s.parseId(id)
	if index >= len(s.values) || s.values[index] == nil || s.values[index].GetArrayKey() != id {
		return true
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.values[index] = nil
	s.remove.Add(index)
	return true
}

//Size 当前socket数量
func (s *Array) Size() int {
	return len(s.values) - s.remove.Size()
}

//遍历
func (s *Array) Range(f func(ArrayVal)) {
	for _, val := range s.values {
		if val != nil {
			f(val)
		}
	}
}
