package utils

import "sync"

type ArraySetKey uint64
type ArraySetVal interface {
	GetArraySetKey() ArraySetKey //获取ArraySet Key
	SetArraySetKey(ArraySetKey)  //设置ArraySet Key
}

func NewArraySet(cap int) *ArraySet {
	arrayMap := &ArraySet{
		seed:   1,
		remove: newArrayIndex(cap),
		values: make([]ArraySetVal, cap, cap),
	}
	for i := cap - 1; i >= 0; i-- {
		arrayMap.remove.Add(i)
	}
	return arrayMap
}

type ArraySet struct {
	seed   uint32 //ID 生成种子
	mutex  sync.Mutex
	values []ArraySetVal
	remove *arrayIndex
}

//createSocketId 使用index生成ID
func (s *ArraySet) createId(index int) ArraySetKey {
	s.seed++
	return ArraySetKey(index)<<32 | ArraySetKey(s.seed)
}

//parseSocketId 返回idPack中的index
func (s *ArraySet) parseId(id ArraySetKey) int {
	return int(id >> 32)
}

func (s *ArraySet) Add(v ArraySetVal) ArraySetKey {
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
	v.SetArraySetKey(id)
	return id
}

//Get 获取
func (s *ArraySet) Get(id ArraySetKey) ArraySetVal {
	index := s.parseId(id)
	if index >= len(s.values) {
		return nil
	}
	if val := s.values[index]; val != nil && val.GetArraySetKey() == id {
		return val
	} else {
		return nil
	}
}

//Delete 删除
func (s *ArraySet) Delete(id ArraySetKey) bool {
	index := s.parseId(id)
	if index >= len(s.values) || s.values[index] == nil || s.values[index].GetArraySetKey() != id {
		return true
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.values[index] = nil
	s.remove.Add(index)
	return true
}

//Size 当前socket数量
func (s *ArraySet) Size() int {
	return len(s.values) - s.remove.Size()
}

//遍历
func (s *ArraySet) Range(f func(ArraySetVal)) {
	for _, val := range s.values {
		if val != nil {
			f(val)
		}
	}
}
