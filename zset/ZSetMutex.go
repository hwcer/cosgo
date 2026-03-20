package zset

import "sync"

// ZSetMutex 是线程安全的 ZSet 包装
type ZSetMutex struct {
	*ZSet
	lock sync.RWMutex
}

// NewZSetMutex 创建线程安全的 ZSet
func NewZSetMutex(order ...int8) *ZSetMutex {
	return &ZSetMutex{
		ZSet: New(order...),
	}
}

// ZAdd 添加或更新元素（线程安全）
func (s *ZSetMutex) ZAdd(score int64, key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ZSet.ZAdd(score, key)
}

// ZIncr 增加分数（线程安全）
func (s *ZSetMutex) ZIncr(score int64, key string) int64 {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.ZSet.ZIncr(score, key)
}

// ZRem 删除元素（线程安全）
func (s *ZSetMutex) ZRem(key string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.ZSet.ZRem(key)
}

// ZRank 获取排名（线程安全）
func (s *ZSetMutex) ZRank(key string) (int64, int64) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.ZSet.ZRank(key)
}

// ZScore 获取分数（线程安全）
func (s *ZSetMutex) ZScore(key string) (int64, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.ZSet.ZScore(key)
}

// ZData 获取指定排名的元素（线程安全）
func (s *ZSetMutex) ZData(rank int64) (string, int64) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.ZSet.ZData(rank)
}

// ZRange 范围遍历（线程安全）
func (s *ZSetMutex) ZRange(start, end int64, f func(int64, string)) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	s.ZSet.ZRange(start, end, f)
}

// ZCount 计数（线程安全）
func (s *ZSetMutex) ZCount(min, max int64) int64 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.ZSet.ZCount(min, max)
}

// GetGuardScore 获取守门员分数（线程安全）
func (s *ZSetMutex) GetGuardScore() (int64, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.ZSet.GetGuardScore()
}

// IsFull 检查是否满员（线程安全）
func (s *ZSetMutex) IsFull() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.ZSet.IsFull()
}

// SetMaxSize 设置人数限制（线程安全）
func (s *ZSetMutex) SetMaxSize(maxSize int32) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ZSet.SetMaxSize(maxSize)
}

// GetMaxSize 获取人数限制（线程安全）
func (s *ZSetMutex) GetMaxSize() int32 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.ZSet.GetMaxSize()
}

// ZCard 获取元素数量（线程安全）
func (s *ZSetMutex) ZCard() int64 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.ZSet.ZCard()
}
