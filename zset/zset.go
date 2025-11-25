package zset

import (
	"sync"
)

// ZSet 是针对string类型key和int64类型分数的有序集合实现
type ZSet struct {
	dict map[string]int64
	zsl  *skipList
	lock sync.RWMutex
}

/*-----------------------------------------------------------------------------
 * Common sorted set API
 *----------------------------------------------------------------------------*/

// New 创建一个新的ZSet并返回其指针
func New(order ...int8) *ZSet {
	var zOrder int8
	if len(order) > 0 {
		zOrder = order[0]
	}
	s := &ZSet{
		dict: make(map[string]int64),
		zsl:  zslCreate(zOrder),
		lock: sync.RWMutex{},
	}
	return s
}

// ZCard 返回元素数量
func (z *ZSet) ZCard() int64 {
	z.lock.RLock()
	defer z.lock.RUnlock()
	return z.zsl.length
}

// zadd 内部方法，用于添加或更新元素，返回最终的分数
func (z *ZSet) zadd(score int64, key string) int64 {
	// 首先检查key是否已存在
	oldScore, exists := z.dict[key]

	// 如果key已存在且分数相同，则不需要修改
	if exists && score == oldScore {
		return score
	}

	// 更新字典中的分数

	// 如果key已存在但分数不同，则先删除再重新插入
	if exists {
		success := z.zsl.zslDelete(oldScore, key)
		// 确保删除成功才进行插入
		if success {
			z.zsl.zslInsert(score, key)
		} else {
			// 如果删除失败，说明可能存在数据不一致，启用强制删除
			//logger.Trace("ZAdd failed, delete old score failed, key: %s, score: %d. Attempting force delete.", key, oldScore)
			forceDeleted := z.zsl.zslForceDeleteById(key)
			if forceDeleted > 0 {
				// 强制删除成功后插入新的分数
				//logger.Trace("Force delete successful for key: %s, deleted %d nodes", key, forceDeleted)
				z.zsl.zslInsert(score, key)
			} else {
				// 强制删除也失败
				return score
				//logger.Trace("Force delete also failed for key: %s. Rolling back score update.", key)
				//z.dict[key] = oldScore
			}
		}
	} else {
		// 对于新key，直接插入
		z.zsl.zslInsert(score, key)
	}

	z.dict[key] = score
	return score
}

// ZAdd 用于添加或更新元素
func (z *ZSet) ZAdd(score int64, key string) {
	z.lock.Lock()
	defer z.lock.Unlock()
	z.zadd(score, key)
}

// ZIncr 有序集合中对指定成员的分数加上增量 score
func (z *ZSet) ZIncr(score int64, key string) int64 {
	z.lock.Lock()
	defer z.lock.Unlock()
	if score == 0 {
		// 如果增量为0，直接返回当前分数
		if currentScore, ok := z.dict[key]; ok {
			return currentScore
		}
		return 0
	}
	newScore := z.dict[key] + score
	return z.zadd(newScore, key)
}

// ZRem 通过key从ZSet中删除元素
func (z *ZSet) ZRem(key string) (ok bool) {
	z.lock.Lock()
	defer z.lock.Unlock()
	score, ok := z.dict[key]
	if ok {
		// 先验证跳表删除是否成功
		deleted := z.zsl.zslDelete(score, key)
		if deleted {
			delete(z.dict, key)
			return true
		}
		// 如果跳表删除失败，返回false
		return false
	}
	return false
}

// ZRank 返回元素的正序排名（从0开始）和分数
func (z *ZSet) ZRank(key string) (rank int64, score int64) {
	return z.zRank(key, false)
}

// ZRevRank 返回元素的逆序排名（从0开始）和分数
func (z *ZSet) ZRevRank(key string) (rank int64, score int64) {
	return z.zRank(key, true)
}

// zRank 内部方法，用于获取元素的排名和分数
// 参数reverse决定排名是降序还是升序，true表示降序，false表示升序
func (z *ZSet) zRank(key string, reverse bool) (rank int64, score int64) {
	z.lock.RLock()
	defer z.lock.RUnlock()
	score, ok := z.dict[key]
	if !ok {
		return -1, 0
	}
	r := z.zsl.zslGetRank(score, key)
	if reverse {
		r = z.zsl.length - r
	} else {
		r--
	}
	return r, score
}

// ZScore 通过key获取分数
func (z *ZSet) ZScore(key string) (score int64, ok bool) {
	z.lock.RLock()
	defer z.lock.RUnlock()
	score, ok = z.dict[key]
	return score, ok
}

// ZData 返回指定排名的元素的key和分数（正序）
func (z *ZSet) ZData(rank int64) (key string, score int64) {
	return z.zData(rank, false)
}

// ZRevData 返回指定排名的元素的key和分数（逆序）
func (z *ZSet) ZRevData(rank int64) (key string, score int64) {
	return z.zData(rank, true)
}

// zData 内部方法，用于获取指定排名的元素的key和分数
// 参数rank是排名位置，reverse表示是否为逆序排名
func (z *ZSet) zData(rank int64, reverse bool) (key string, score int64) {
	z.lock.RLock()
	defer z.lock.RUnlock()
	if rank < 0 || rank >= z.zsl.length {
		return "", 0
	}
	if reverse {
		rank = z.zsl.length - rank - 1
	}
	n := z.zsl.zslGetElementByRank(uint64(rank + 1))
	if n == nil {
		return "", 0
	}
	score, ok := z.dict[n.id]
	if !ok {
		return "", 0
	}
	return n.id, score
}

// ZRange 实现ZRANGE命令，按照分数升序遍历指定范围的元素
func (z *ZSet) ZRange(start, end int64, f func(int64, string)) {
	z.snapshotRange(start, end, false, f)
}

// ZRevRange 实现ZREVRANGE命令，按照分数降序遍历指定范围的元素
func (z *ZSet) ZRevRange(start, end int64, f func(int64, string)) {
	z.snapshotRange(start, end, true, f)
}

// snapshotRange 内部方法，先获取元素快照，然后在锁外执行回调函数
func (z *ZSet) snapshotRange(start, end int64, reverse bool, f func(int64, string)) {
	scores := make([]int64, 0)
	keys := make([]string, 0)

	z.lock.RLock()
	z.commonRange(start, end, reverse, func(f int64, k string) {
		scores = append(scores, f)
		keys = append(keys, k)
	})
	z.lock.RUnlock()

	for i, score := range scores {
		f(score, keys[i])
	}
}

// commonRange 内部方法，实现范围查询的通用逻辑
func (z *ZSet) commonRange(start, end int64, reverse bool, f func(int64, string)) {
	l := z.zsl.length
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
		return
	}
	if end >= l {
		end = l - 1
	}
	span := (end - start) + 1

	var node *zNode
	if reverse {
		node = z.zsl.tail
		if start > 0 {
			node = z.zsl.zslGetElementByRank(uint64(l - start))
		}
	} else {
		node = z.zsl.header.level[0].forward
		if start > 0 {
			node = z.zsl.zslGetElementByRank(uint64(start + 1))
		}
	}
	for span > 0 {
		span--
		k := node.id
		s := node.score
		f(s, k)
		if reverse {
			node = node.backward
		} else {
			node = node.level[0].forward
		}
	}
}

// ZCount 返回有序集合中分数在min和max之间的元素数量
func (z *ZSet) ZCount(min, max int64) int64 {
	z.lock.RLock()
	defer z.lock.RUnlock()
	spec := &zRangeSpec{
		min:   min,
		max:   max,
		minex: 0,
		maxex: 0,
	}

	// 直接遍历范围内的元素计数
	var count int64 = 0
	for x := z.zsl.header.level[0].forward; x != nil; x = x.level[0].forward {
		if zslValueGteMin(x.score, spec) && zslValueLteMax(x.score, spec) {
			count++
		} else if x.score > max {
			// 因为是有序的，所以当分数超过max时可以提前退出
			break
		}
	}
	return count
}
