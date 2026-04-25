// ZSet 有序集合
//
// 基于「跳表 + 字典」双维护实现，适合实时排行榜、Top-N 统计等场景。
//
// 核心规则：
//  1. 排序：默认降序（高分在前），可选升序；同分按插入顺序排列（FIFO，先到先得）
//  2. 守门员：设置 maxSize 后，满员时最后一名为守门员，新成员分数必须严格优于守门员才能入榜
//  3. 懒淘汰：超出 maxSize 的冗余数据不立即清理，累积到阈值后批量裁剪，以空间换吞吐
package zset

import (
	"sync"
	"sync/atomic"
)

const (
	// cleanupBufferSize 懒淘汰阈值
	cleanupBufferSize = 100
)

// ZSet 有序集合主结构
type ZSet struct {
	dict    map[string]int64 // key → score 字典，O(1) 查分数
	zsl     *skipList        // 跳��，维护排名顺序
	maxSize int32            // 排行榜人数上限，0 表示不限制
	guard   guardKeeper      // 守门员（锁内使用）
	lock    sync.RWMutex     // 读写锁

	// 守门员分数的原子副本，供 CanEnter 无锁预检使用
	// 与 guard 字段保持同步，在 updateGuard 中同时写入
	guardScore atomic.Int64
	guardValid atomic.Int32 // 0=无效，1=有效（用 Int32 代替 atomic.Bool 以兼容旧版 Go）
}

// ZNode 范围查询返回的节点
type ZNode struct {
	Key   string
	Rank  int64
	Score int64
}

// guardKeeper 守门员结构（锁内使用）
type guardKeeper struct {
	score int64
	key   string
	valid bool
}

// New 创建有序集合
// order < 0 降序（默认），order > 0 升序
func New(order ...int8) *ZSet {
	var zOrder int8
	if len(order) > 0 {
		zOrder = order[0]
	} else {
		zOrder = -1
	}
	return &ZSet{
		dict: make(map[string]int64),
		zsl:  zslCreate(zOrder),
	}
}

// NewWithMaxSize 创建带人数限制的有序集合
func NewWithMaxSize(maxSize int32, order ...int8) *ZSet {
	s := New(order...)
	s.maxSize = maxSize
	return s
}

// CanEnter 无锁预检：判断指定分数是否有可能入榜
// 使用 atomic 读取守门员分数，不获取任何锁
//
// 用途：调用方在业务层做快速过滤，避免无效的 ZAdd 写锁竞争
//
//	if z.CanEnter(score) {
//	    z.ZAdd(score, key)
//	}
//
// 注意：存在极小的竞态窗口（守门员刚被更新），可能出现假阳性（返回 true 但实际被拒）
// 但不会出现假阴性（返回 false 但实际能入），因此是安全的预过滤
func (z *ZSet) CanEnter(score int64) bool {
	if z.maxSize <= 0 {
		return true
	}
	if z.guardValid.Load() == 0 {
		return true
	}
	gs := z.guardScore.Load()
	if z.zsl.order < 0 {
		return score > gs
	}
	return score < gs
}

// canEnter 内部版本（需在锁内调用）
func (z *ZSet) canEnter(score int64) bool {
	if z.maxSize <= 0 {
		return true
	}
	if z.zsl.length < int64(z.maxSize) {
		return true
	}
	if !z.guard.valid {
		return true
	}
	return z.compareScores(score, z.guard.score) > 0
}

// upsert 插入或更新元素的内部实现，由 ZAdd 和 ZIncr 共用
func (z *ZSet) upsert(key string, newScore int64, exists bool, oldScore int64) int64 {
	if exists && newScore == oldScore {
		return newScore
	}

	canEnter := z.canEnter(newScore)
	wasFull := z.maxSize > 0 && z.zsl.length >= int64(z.maxSize)
	isGuard := z.guard.valid && key == z.guard.key

	if exists {
		z.dict[key] = newScore
		wasInSkiplist := z.zsl.zslDelete(oldScore, key)
		if canEnter || wasInSkiplist {
			z.zsl.zslInsert(newScore, key)
		}
	} else if canEnter {
		z.dict[key] = newScore
		z.zsl.zslInsert(newScore, key)
	} else {
		return 0
	}

	if z.maxSize > 0 {
		needUpdateGuard := false
		if !wasFull && z.zsl.length >= int64(z.maxSize) {
			needUpdateGuard = true
		}
		if wasFull && (isGuard || (!exists && canEnter)) {
			needUpdateGuard = true
		}
		if needUpdateGuard {
			z.updateGuard()
		}
	}

	z.tryTrimExcess()
	return newScore
}

// ZAdd 添加或更新元素，返回最终分数；被守门员拦截返回 0
func (z *ZSet) ZAdd(score int64, key string) int64 {
	z.lock.Lock()
	defer z.lock.Unlock()
	oldScore, exists := z.dict[key]
	return z.upsert(key, score, exists, oldScore)
}

// ZIncr 对指定元素的分数加上增量
func (z *ZSet) ZIncr(score int64, key string) int64 {
	z.lock.Lock()
	defer z.lock.Unlock()
	if score == 0 {
		if currentScore, ok := z.dict[key]; ok {
			return currentScore
		}
		return 0
	}
	oldScore, exists := z.dict[key]
	newScore := oldScore + score
	return z.upsert(key, newScore, exists, oldScore)
}

// ZRem 删除指定 key 的元素
func (z *ZSet) ZRem(key string) bool {
	z.lock.Lock()
	defer z.lock.Unlock()
	score, ok := z.dict[key]
	if !ok {
		return false
	}

	deleted := z.zsl.zslDelete(score, key)
	delete(z.dict, key)

	if deleted && z.guard.valid {
		if key == z.guard.key || z.compareScores(score, z.guard.score) == 1 {
			z.updateGuard()
		}
	}

	return true
}

// ZRemRangeByRank 删除排名在 [start, stop] 范围内的元素（0-based，支持负数索引）
func (z *ZSet) ZRemRangeByRank(start, stop int64) int64 {
	z.lock.Lock()
	defer z.lock.Unlock()
	l := z.zcard()

	if start < 0 {
		start += l
		if start < 0 {
			start = 0
		}
	}
	if stop < 0 {
		stop += l
	}
	if start > stop || start >= l {
		return 0
	}
	if stop >= l {
		stop = l - 1
	}

	removed := z.zsl.zslDeleteRangeByRank(start, stop, z.dict)

	if removed > 0 && z.guard.valid {
		z.updateGuard()
	}

	return removed
}

// ZRemRangeByScore 删除分数在 [min, max] 范围内的元素
func (z *ZSet) ZRemRangeByScore(min, max int64) int64 {
	z.lock.Lock()
	defer z.lock.Unlock()

	if z.guard.valid {
		if z.zsl.order < 0 {
			if min < z.guard.score {
				min = z.guard.score
			}
		} else {
			if max > z.guard.score {
				max = z.guard.score
			}
		}
	}

	removed := z.zsl.zslDeleteRangeByScore(min, max, z.dict)

	if removed > 0 && z.guard.valid {
		z.updateGuard()
	}

	return removed
}

// ZRank 获取元素排名（0-based）和分数，不存在或未入榜返回 rank=-1
func (z *ZSet) ZRank(key string) (rank int64, score int64) {
	z.lock.RLock()
	defer z.lock.RUnlock()
	score, ok := z.dict[key]
	if !ok {
		return -1, 0
	}

	if z.guard.valid {
		if z.compareScores(score, z.guard.score) < 0 {
			return -1, 0
		}
	}

	return z.zsl.zslRank(score, key), score
}

// ZScore 获取元素分数，即使未入榜也能查到
func (z *ZSet) ZScore(key string) (score int64, ok bool) {
	z.lock.RLock()
	defer z.lock.RUnlock()
	score, ok = z.dict[key]
	return score, ok
}

// ZElement 获取指定排名的元素
func (z *ZSet) ZElement(rank int64) (key string, score int64) {
	z.lock.RLock()
	defer z.lock.RUnlock()
	if rank < 0 || rank >= z.zcard() {
		return "", 0
	}
	n := z.zsl.zslElement(rank)
	if n == nil {
		return "", 0
	}
	return n.id, n.score
}

// ZRange 按排名范围获取元素列表（0-based，支持负数索引）
func (z *ZSet) ZRange(start, end int64) []ZNode {
	z.lock.RLock()
	defer z.lock.RUnlock()
	l := z.zcard()

	if start < 0 {
		start += l
		if start < 0 {
			start = 0
		}
	}
	if end < 0 {
		end += l
	}
	if start > end || start >= l {
		return nil
	}
	if end >= l {
		end = l - 1
	}

	return z.zsl.zslRange(start, end)
}

// ZRangeWithCallback 按排名范围遍历元素
func (z *ZSet) ZRangeWithCallback(start, end int64, f func(int64, string)) {
	nodes := z.ZRange(start, end)
	for _, node := range nodes {
		f(node.Score, node.Key)
	}
}

// ZRangeByScore 按分数范围获取元素列表
func (z *ZSet) ZRangeByScore(min, max int64) []ZNode {
	z.lock.RLock()
	defer z.lock.RUnlock()

	if z.guard.valid {
		if z.zsl.order < 0 {
			if min < z.guard.score {
				min = z.guard.score
			}
		} else {
			if max > z.guard.score {
				max = z.guard.score
			}
		}
	}

	return z.zsl.zslRangeByScore(min, max)
}

// ZRangeByScoreWithCallback 按分数范围遍历元素
func (z *ZSet) ZRangeByScoreWithCallback(min, max int64, f func(int64, string)) {
	nodes := z.ZRangeByScore(min, max)
	for _, node := range nodes {
		f(node.Score, node.Key)
	}
}

// compareScores 比较两个分数的排名位置关系
// 返回 1（score1 更优）、0（相等）、-1（score1 更差）
func (z *ZSet) compareScores(score1, score2 int64) int {
	if score1 == score2 {
		return 0
	}
	if z.zsl.order < 0 {
		if score1 > score2 {
			return 1
		}
		return -1
	}
	if score1 < score2 {
		return 1
	}
	return -1
}

// ZCount 统计分数在 [min, max] 范围内的元素数量，O(log n)
func (z *ZSet) ZCount(min, max int64) int64 {
	z.lock.RLock()
	defer z.lock.RUnlock()

	if z.guard.valid {
		if z.zsl.order < 0 {
			if min < z.guard.score {
				min = z.guard.score
			}
		} else {
			if max > z.guard.score {
				max = z.guard.score
			}
		}
	}

	return z.zsl.zslCount(min, max)
}

// ZCard 返回排行榜有效元素个数
func (z *ZSet) ZCard() int64 {
	z.lock.RLock()
	defer z.lock.RUnlock()
	return z.zcard()
}

// zcard 内部无锁版本
func (z *ZSet) zcard() int64 {
	if z.maxSize > 0 && z.zsl.length > int64(z.maxSize) {
		return int64(z.maxSize)
	}
	return z.zsl.length
}

// updateGuard 更新守门员，同时同步原子副本供 CanEnter 无锁预检使用
func (z *ZSet) updateGuard() {
	if z.maxSize <= 0 || z.zsl.length < int64(z.maxSize) {
		z.guard.valid = false
		z.guardValid.Store(0)
		return
	}

	n := z.zsl.zslElement(int64(z.maxSize) - 1)
	if n == nil {
		z.guard.valid = false
		z.guardValid.Store(0)
		return
	}

	z.guard.score = n.score
	z.guard.key = n.id
	z.guard.valid = true

	// 同步到原子字段，供 CanEnter 无锁读取
	z.guardScore.Store(n.score)
	z.guardValid.Store(1)
}

// tryTrimExcess 懒淘汰：当 dict 超出 maxSize + 阈值时批量清理跳表末尾
func (z *ZSet) tryTrimExcess() {
	if z.maxSize <= 0 {
		return
	}

	currentSize := int64(len(z.dict))
	maxSize := int64(z.maxSize)

	if currentSize-maxSize < cleanupBufferSize {
		return
	}

	if z.zsl.length > int64(z.maxSize) {
		z.zsl.zslDeleteRangeByRank(int64(z.maxSize), z.zsl.length-1, z.dict)
	}
}
