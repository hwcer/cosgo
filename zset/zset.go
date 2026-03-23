package zset

import (
	"sync"
)

const (
	// cleanupBufferSize 清理缓冲大小
	// 当超出人数限制超过此值时才触发清理
	cleanupBufferSize = 100
)

// ZSet 是针对 string 类型 key 和 int64 类型分数的有序集合实现
type ZSet struct {
	dict    map[string]int64
	zsl     *skipList
	maxSize int32       // 最大人数限制，0=不限制
	guard   guardKeeper // 守门员（最后一名分数）
	lock    sync.RWMutex
}

// ZNode 表示有序集合中的一个节点
type ZNode struct {
	Key   string
	Rank  int64
	Score int64
}

// guardKeeper 守门员结构
type guardKeeper struct {
	score int64  // 最后一名的分数
	key   string // 最后一名的 key
	valid bool   // 是否有效（满员时才有效）
}

// New 创建一个新的 ZSet 并返回其指针
func New(order ...int8) *ZSet {
	var zOrder int8
	if len(order) > 0 {
		zOrder = order[0]
	} else {
		zOrder = -1 // 默认降序
	}
	s := &ZSet{
		dict:    make(map[string]int64),
		zsl:     zslCreate(zOrder),
		maxSize: 0, // 默认不限制
		guard:   guardKeeper{valid: false},
	}
	return s
}

// NewWithMaxSize 创建带人数限制的 ZSet
func NewWithMaxSize(maxSize int32, order ...int8) *ZSet {
	s := New(order...)
	s.maxSize = maxSize
	return s
}

// canEnter 检查是否能进入排行榜
func (z *ZSet) canEnter(score int64) bool {
	// 未限制人数
	if z.maxSize <= 0 {
		return true
	}

	// 跳表未满员，允许
	if z.zsl.length < int64(z.maxSize) {
		return true
	}

	// 跳表已满，需要和守门员比较
	if !z.guard.valid {
		return true // 守门员无效，允许
	}

	// 使用 compareScores 方法判断是否能进入
	// compareScores 返回 1 表示 score 排在 guard.score 前面
	return z.compareScores(score, z.guard.score) > 0
}

// zadd 内部方法，用于添加或更新元素，返回最终的分数
func (z *ZSet) ZAdd(score int64, key string) int64 {
	z.lock.Lock()
	defer z.lock.Unlock()
	// 首先检查 key 是否已存在
	oldScore, exists := z.dict[key]

	// 如果 key 已存在且分数相同，则不需要修改
	if exists && score == oldScore {
		return score
	}

	// 统一检查是否能进入排行榜（只调用一次）
	canEnter := z.canEnter(score)

	// 记录插入前的状态，用于智能更新守门员
	wasFull := z.maxSize > 0 && z.zsl.length >= int64(z.maxSize)
	isGuard := z.guard.valid && key == z.guard.key

	// 1. 如果已存在，先更新 dict（无论是否在榜单内）
	if exists {
		z.dict[key] = score

		// 已存在元素：尝试从跳表中删除旧位置
		// 返回值表示是否真的在跳表中
		wasInSkiplist := z.zsl.zslDelete(oldScore, key)

		// 决定是否需要重新插入跳表
		if canEnter || wasInSkiplist {
			// 允许进入，或者之前在跳表中（懒淘汰：保持连续性）
			z.zsl.zslInsert(score, key)
		}
	} else if canEnter {
		// 2. 如果是新元素且允许进入，更新 dict 并插入跳表
		z.dict[key] = score
		z.zsl.zslInsert(score, key)
	} else {
		// 3. 新元素且不允许进入，直接拒绝
		return 0
	}

	// 4. 智能更新守门员（只在可能改变时才更新）
	if z.maxSize > 0 {
		needUpdateGuard := false

		// 情况 1：从未满到满
		if !wasFull && z.zsl.length >= int64(z.maxSize) {
			needUpdateGuard = true
		}

		// 情况 2：已经是满的，且操作影响了守门员位置
		if wasFull {
			// 删除了守门员或更新了守门员的分数
			if isGuard {
				needUpdateGuard = true
			}

			// 新元素插入可能改变了守门员位置
			if !exists && canEnter {
				needUpdateGuard = true
			}
		}

		if needUpdateGuard {
			z.updateGuard()
		}
	}

	// 5. 检查是否需要裁剪排行（超员超过阈值才触发）
	z.tryTrimExcess()

	return score
}

// ZIncr 有序集合中对指定成员的分数加上增量 score
func (z *ZSet) ZIncr(score int64, key string) int64 {
	z.lock.Lock()
	defer z.lock.Unlock()
	if score == 0 {
		// 如果增量为 0，直接返回当前分数
		if currentScore, ok := z.dict[key]; ok {
			return currentScore
		}
		return 0
	}
	newScore := z.dict[key] + score
	// 首先检查 key 是否已存在
	oldScore, exists := z.dict[key]

	// 统一检查是否能进入排行榜（只调用一次）
	canEnter := z.canEnter(newScore)

	// 记录插入前的状态，用于智能更新守门员
	wasFull := z.maxSize > 0 && z.zsl.length >= int64(z.maxSize)
	isGuard := z.guard.valid && key == z.guard.key

	// 1. 如果已存在，先更新 dict（无论是否在榜单内）
	if exists {
		z.dict[key] = newScore

		// 已存在元素：尝试从跳表中删除旧位置
		// 返回值表示是否真的在跳表中
		wasInSkiplist := z.zsl.zslDelete(oldScore, key)

		// 决定是否需要重新插入跳表
		if canEnter || wasInSkiplist {
			// 允许进入，或者之前在跳表中（懒淘汰：保持连续性）
			z.zsl.zslInsert(newScore, key)
		}
	} else if canEnter {
		// 2. 如果是新元素且允许进入，更新 dict 并插入跳表
		z.dict[key] = newScore
		z.zsl.zslInsert(newScore, key)
	} else {
		// 3. 新元素且不允许进入，直接拒绝
		return 0
	}

	// 4. 智能更新守门员（只在可能改变时才更新）
	if z.maxSize > 0 {
		needUpdateGuard := false

		// 情况 1：从未满到满
		if !wasFull && z.zsl.length >= int64(z.maxSize) {
			needUpdateGuard = true
		}

		// 情况 2：已经是满的，且操作影响了守门员位置
		if wasFull {
			// 删除了守门员或更新了守门员的分数
			if isGuard {
				needUpdateGuard = true
			}

			// 新元素插入可能改变了守门员位置
			if !exists && canEnter {
				needUpdateGuard = true
			}
		}

		if needUpdateGuard {
			z.updateGuard()
		}
	}

	// 5. 检查是否需要裁剪排行（超员超过阈值才触发）
	z.tryTrimExcess()

	return newScore
}

// ZRem 通过 key 从 ZSet 中删除元素
func (z *ZSet) ZRem(key string) bool {
	z.lock.Lock()
	defer z.lock.Unlock()
	score, ok := z.dict[key]
	if !ok {
		return false
	}

	// 先验证跳表删除是否成功
	deleted := z.zsl.zslDelete(score, key)
	if deleted {
		delete(z.dict, key)

		// 如果删除的是守门员或者守门员前面的元素，需要重新计算
		if z.guard.valid {
			if key == z.guard.key || z.compareScores(score, z.guard.score) == 1 {
				z.updateGuard()
			}
		}

		return true
	}

	// 跳表删除失败，可能是在守门员外的元素
	// 只删除 dict
	delete(z.dict, key)
	return true
}

// ZRemRangeByRank 删除指定排名范围的元素（Redis兼容接口）
// start 和 stop 都是从0开始的排名
func (z *ZSet) ZRemRangeByRank(start, stop int64) int64 {
	z.lock.Lock()
	defer z.lock.Unlock()
	l := z.ZCard()

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

	// 调用跳表的删除方法
	removed := z.zsl.zslDeleteRangeByRank(start, stop, z.dict)

	// 删除后可能需要更新守门员
	if removed > 0 && z.guard.valid {
		z.updateGuard()
	}

	return removed
}

// ZRemRangeByScore 删除指定分数范围的元素（Redis兼容接口）
func (z *ZSet) ZRemRangeByScore(min, max int64) int64 {
	z.lock.Lock()
	defer z.lock.Unlock()
	// 考虑守门员限制
	if z.guard.valid {
		// 根据排序方式调整删除范围，确保不越过守门员
		if z.zsl.order < 0 {
			// 降序：守门员是分数最低的，不能删除分数低于守门员的元素
			if min < z.guard.score {
				min = z.guard.score
			}
		} else {
			// 升序：守门员是分数最高的，不能删除分数高于守门员的元素
			if max > z.guard.score {
				max = z.guard.score
			}
		}
	}

	// 调用跳表的删除方法
	removed := z.zsl.zslDeleteRangeByScore(min, max, z.dict)

	// 删除后可能需要更新守门员
	if removed > 0 && z.guard.valid {
		z.updateGuard()
	}

	return removed
}

// ZRank 返回元素的排名（从 0 开始）和分数
// 如果元素在守门员之外或不存在，返回 -1
func (z *ZSet) ZRank(key string) (rank int64, score int64) {
	z.lock.RLock()
	defer z.lock.RUnlock()
	score, ok := z.dict[key]
	if !ok {
		return -1, 0
	}

	// 如果设置了守门员，检查元素是否在守门员之内
	if z.guard.valid {
		// 使用 compareScores 判断元素是否在守门员之前或等于守门员
		// compareScores 返回 1 表示 score 排在 guard.score 前面
		// 返回 0 表示相等，返回 -1 表示在后面
		cmp := z.compareScores(score, z.guard.score)
		if cmp < 0 {
			// 元素在守门员之后，返回 -1
			return -1, 0
		}
	}

	return z.zsl.zslRank(score, key), score
}

// ZScore 通过 key 获取分数
func (z *ZSet) ZScore(key string) (score int64, ok bool) {
	z.lock.RLock()
	defer z.lock.RUnlock()
	score, ok = z.dict[key]
	return score, ok
}

// ZElement 返回指定排名的元素的 key 和分数
func (z *ZSet) ZElement(rank int64) (key string, score int64) {
	z.lock.RLock()
	defer z.lock.RUnlock()
	if rank < 0 || rank >= z.ZCard() {
		return "", 0
	}

	// zslElement 从0开始计数
	n := z.zsl.zslElement(rank)
	if n == nil {
		return "", 0
	}

	score, ok := z.dict[n.id]
	if !ok {
		return "", 0
	}

	return n.id, score
}

// ZRange 按照跳表顺序返回指定范围的元素
func (z *ZSet) ZRange(start, end int64) []ZNode {
	z.lock.RLock()
	defer z.lock.RUnlock()
	// 使用 ZCard() 获取元素个数，确保与 ZCard 方法保持一致
	l := z.ZCard()

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

	// 调用跳表的范围查询方法
	return z.zsl.zslRange(start, end)
}

// ZRange 按照跳表顺序遍历指定范围的元素（带回调）
func (z *ZSet) ZRangeWithCallback(start, end int64, f func(int64, string)) {
	nodes := z.ZRange(start, end)
	for _, node := range nodes {
		f(node.Score, node.Key)
	}
}

// ZRangeByScore 按照分数范围返回元素节点
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

// ZRangeByScoreWithCallback 按照分数范围遍历元素（带回调）
func (z *ZSet) ZRangeByScoreWithCallback(min, max int64, f func(int64, string)) {
	nodes := z.ZRangeByScore(min, max)

	for _, node := range nodes {
		f(node.Score, node.Key)
	}
}

// compareScores 比较两个分数的位置关系，考虑排序方向
// 返回值：
//
//	1: score1 在 score2 前面
//	0: score1 和 score2 相等
//
// -1: score1 在 score2 后面
func (z *ZSet) compareScores(score1, score2 int64) int {
	if score1 == score2 {
		return 0
	}

	if z.zsl.order < 0 {
		// 降序：分数大的在前面
		if score1 > score2 {
			return 1
		} else {
			return -1
		}
	} else {
		// 升序：分数小的在前面
		if score1 < score2 {
			return 1
		} else {
			return -1
		}
	}
}

// ZCount 返回有序集合中分数在 min 和 max 之间的元素数量
func (z *ZSet) ZCount(min, max int64) int64 {
	z.lock.RLock()
	defer z.lock.RUnlock()
	// 考虑守门员限制
	if z.guard.valid {
		// 根据排序方式调整统计范围，确保不越过守门员
		if z.zsl.order < 0 {
			// 降序：守门员是分数最低的，不能统计分数低于守门员的元素
			if min < z.guard.score {
				min = z.guard.score
			}
		} else {
			// 升序：守门员是分数最高的，不能统计分数高于守门员的元素
			if max > z.guard.score {
				max = z.guard.score
			}
		}
	}

	// 使用跳表的范围计数方法
	count := z.zsl.zslCount(min, max)

	return count
}

// ZCard 返回有序集合的元素个数
func (z *ZSet) ZCard() int64 {
	z.lock.RLock()
	defer z.lock.RUnlock()
	if z.maxSize > 0 && z.zsl.length > int64(z.maxSize) {
		// 懒裁剪模式下，实际元素个数不应该超过 maxSize
		return int64(z.maxSize)
	}
	return z.zsl.length
}

// updateGuard 更新守门员
func (z *ZSet) updateGuard() {
	// 只在跳表满员时才需要守门员
	if z.maxSize <= 0 || z.zsl.length < int64(z.maxSize) {
		z.guard.valid = false
		return
	}

	// 获取第 maxSize 名（跳表的最后一名）
	rank := int64(z.maxSize) - 1
	key, score := z.ZElement(rank)
	if key == "" {
		z.guard.valid = false
		return
	}

	z.guard.score = score
	z.guard.key = key
	z.guard.valid = true
}

// tryTrimExcess 尝试清理超出人数限制的成员（带常数阈值）
func (z *ZSet) tryTrimExcess() {
	if z.maxSize <= 0 {
		return
	}

	currentSize := int64(len(z.dict))
	maxSize := int64(z.maxSize)

	// 只有超出阈值才清理
	if currentSize-maxSize < cleanupBufferSize {
		return
	}

	// 清理跳表中超出 maxSize 的元素
	if z.zsl.length > int64(z.maxSize) {
		// 删除从 maxSize 到末尾的元素（从0开始计数）
		// zslDeleteRangeByRank 会返回成功删除的元素数量，并在内部自动清理 dict
		// 清理的是守门员后面的元素，不会影响到守门员本身
		z.zsl.zslDeleteRangeByRank(int64(z.maxSize), z.zsl.length-1, z.dict)
	}
}
