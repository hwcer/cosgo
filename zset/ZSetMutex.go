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

// NewZSetMutexWithMaxSize 创建带人数限制的线程安全 ZSet
func NewZSetMutexWithMaxSize(maxSize int32, order ...int8) *ZSetMutex {
	return &ZSetMutex{
		ZSet: NewWithMaxSize(maxSize, order...),
	}
}

// ZAdd 添加或更新元素（线程安全）
func (s *ZSetMutex) ZAdd(score int64, key string) int64 {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.ZSet.ZAdd(score, key)
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

// ZElement 获取指定排名的元素（线程安全）
func (s *ZSetMutex) ZElement(rank int64) (string, int64) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.ZSet.ZElement(rank)
}

// ZRange 范围遍历（线程安全）
func (s *ZSetMutex) ZRange(start, end int64, f func(int64, string)) {
	// 先获取结果集，然后释放锁
	s.lock.RLock()
	nodes := s.ZSet.ZRange(start, end)
	s.lock.RUnlock()

	// 释放锁后再执行回调函数
	for _, node := range nodes {
		f(node.Score, node.Key)
	}
}

// ZRangeByScore 按照分数范围返回元素节点（线程安全）
func (s *ZSetMutex) ZRangeByScore(min, max int64) []ZNode {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.ZSet.ZRangeByScore(min, max)
}

// ZRangeByScoreWithCallback 按照分数范围遍历（线程安全）
func (s *ZSetMutex) ZRangeByScoreWithCallback(min, max int64, f func(int64, string)) {
	// 先获取结果集，然后释放锁
	s.lock.RLock()
	nodes := s.ZSet.ZRangeByScore(min, max)
	s.lock.RUnlock()

	// 释放锁后再执行回调函数
	for _, node := range nodes {
		f(node.Score, node.Key)
	}
}

// ZRemRangeByRank 删除指定排名范围的元素（线程安全）
func (s *ZSetMutex) ZRemRangeByRank(start, stop int64) int64 {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.ZSet.ZRemRangeByRank(start, stop)
}

// ZRemRangeByScore 删除指定分数范围的元素（线程安全）
func (s *ZSetMutex) ZRemRangeByScore(min, max int64) int64 {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.ZSet.ZRemRangeByScore(min, max)
}

// ZCount 计数（线程安全）
func (s *ZSetMutex) ZCount(min, max int64) int64 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.ZSet.ZCount(min, max)
}

// ZCard 获取元素数量（线程安全）
func (s *ZSetMutex) ZCard() int64 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.ZSet.ZCard()
}
