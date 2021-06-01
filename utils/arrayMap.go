package utils

import "sync"

func IndexOf(arr []int, tar int) int {
	for k, v := range arr {
		if v == tar {
			return k
		}
	}
	return -1
}

type ArrayMapKey uint64
type ArrayMapVal interface {
	GetArrayMapKey() ArrayMapKey //获取ArrayMap Key
}

func NewArrayMap(cap int) *ArrayMap {
	arrayMap := &ArrayMap{
		seed:   1,
		dirty:  newArrayMapRemoveIndex(cap),
		slices: make([]ArrayMapVal, cap, cap),
	}
	for i := cap - 1; i >= 0; i-- {
		arrayMap.dirty.Add(i)
	}
	return arrayMap
}

func newArrayMapRemoveIndex(cap int) *arrayMapRemoveIndex {
	return &arrayMapRemoveIndex{
		list:  make([]int, 0, cap),
		index: -1,
	}
}

//已经被删除的index
type arrayMapRemoveIndex struct {
	list  []int
	index int
}

func (this *arrayMapRemoveIndex) Add(val int) {
	if val < 0 {
		return
	}
	this.index += 1
	if this.index < len(this.list) {
		this.list[this.index] = val
	} else {
		this.list = append(this.list, val)
	}
}

func (this *arrayMapRemoveIndex) Get() int {
	if this.index < 0 {
		return -1
	}
	val := this.list[this.index]
	this.list[this.index] = -1
	this.index -= 1
	return val
}

func (this *arrayMapRemoveIndex) Size() int {
	return this.index + 1
}

type ArrayMap struct {
	seed   uint32 //ID 生成种子
	mutex  sync.Mutex
	dirty  *arrayMapRemoveIndex
	slices []ArrayMapVal
}

//createSocketId 使用index生成ID
func (s *ArrayMap) createId(index int) ArrayMapKey {
	s.seed++
	return ArrayMapKey(index)<<32 | ArrayMapKey(s.seed)
}

//parseSocketId 返回idPack中的index
func (s *ArrayMap) parseId(id ArrayMapKey) int {
	return int(id >> 32)
}

func (s *ArrayMap) Add(v ArrayMapVal) ArrayMapKey {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	var index = -1
	if index = s.dirty.Get(); index >= 0 {
		s.slices[index] = v
	} else {
		index = len(s.slices)
		s.slices = append(s.slices, v)
	}
	return s.createId(index)
}

//Get 获取
func (s *ArrayMap) Get(id ArrayMapKey) ArrayMapVal {
	index := s.parseId(id)
	if index >= len(s.slices) {
		return nil
	}
	if val := s.slices[index]; val != nil && val.GetArrayMapKey() == id {
		return val
	} else {
		return nil
	}
}

//Remove 删除
func (s *ArrayMap) Remove(id ArrayMapKey) bool {
	index := s.parseId(id)
	if index >= len(s.slices) || s.slices[index] == nil || s.slices[index].GetArrayMapKey() != id {
		return true
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.slices[index] = nil
	s.dirty.Add(index)
	return true
}

//Size 当前socket数量
func (s *ArrayMap) Size() int {
	return len(s.slices) - s.dirty.Size()
}

//遍历
func (s *ArrayMap) Range(f func(ArrayMapVal)) {
	for _, val := range s.slices {
		if val != nil {
			f(val)
		}
	}
}
