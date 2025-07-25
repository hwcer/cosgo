package zset

import (
	"golang.org/x/exp/constraints"
	"sync"
)

// ZSet is the final exported sorted set we can use
type ZSet[K constraints.Ordered] struct {
	dict map[K]float64
	zsl  *skipList[K]
	lock sync.RWMutex
}

/*-----------------------------------------------------------------------------
 * Common sorted set API
 *----------------------------------------------------------------------------*/

// New creates a new ZSet and return its pointer
func New[K constraints.Ordered](order ...int8) *ZSet[K] {
	var zOrder int8
	if len(order) > 0 {
		zOrder = order[0]
	}
	s := &ZSet[K]{
		dict: make(map[K]float64),
		zsl:  zslCreate[K](zOrder),
		lock: sync.RWMutex{},
	}
	return s
}

// ZCard returns counts of elements
func (z *ZSet[K]) ZCard() int64 {
	z.lock.RLock()
	defer z.lock.RUnlock()
	return z.zsl.length
}

// ZAdd is used to add or update an element
func (z *ZSet[K]) ZAdd(score float64, key K) {
	z.lock.Lock()
	defer z.lock.Unlock()
	v, ok := z.dict[key]
	z.dict[key] = score
	if ok {
		/* Remove and re-insert when score changes. */
		if score != v {
			z.zsl.zslDelete(v, key)
			z.zsl.zslInsert(score, key)
		}
	} else {
		z.zsl.zslInsert(score, key)
	}
}

// ZIncr ..
// 有序集合中对指定成员的分数加上增量 score
func (z *ZSet[K]) ZIncr(score float64, key K) float64 {
	z.lock.Lock()
	defer z.lock.Unlock()
	oldScore, ok := z.dict[key]
	if !ok {
		z.ZAdd(score, key)
		return score
	}
	if score != 0 {
		z.zsl.zslDelete(oldScore, key)
		z.dict[key] += score
		z.zsl.zslInsert(z.dict[key], key)
	}
	return z.dict[key]
}

// ZRem removes an element from the ZSet
// by its key.
func (z *ZSet[K]) ZRem(key K) (ok bool) {
	z.lock.Lock()
	defer z.lock.Unlock()
	score, ok := z.dict[key]
	if ok {
		z.zsl.zslDelete(score, key)
		delete(z.dict, key)
		return true
	}
	return false
}

func (z *ZSet[K]) ZRank(key K) (rank int64, score float64) {
	return z.zRank(key, false)
}
func (z *ZSet[K]) ZRevRank(key K) (rank int64, score float64) {
	return z.zRank(key, true)
}

// zRank returns position,score and extra data of an element which
// found by the parameter key.
// The parameter reverse determines the rank is descent or ascend，
// true means descend and false means ascend.
func (z *ZSet[K]) zRank(key K, reverse bool) (rank int64, score float64) {
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

// ZScore implements ZScore
// 通过 key 获取分数
func (z *ZSet[K]) ZScore(key K) (score float64, ok bool) {
	z.lock.RLock()
	defer z.lock.RUnlock()
	score, ok = z.dict[key]
	return score, ok
}

func (z *ZSet[K]) ZData(rank int64) (key K, score float64) {
	return z.zData(rank, false)
}
func (z *ZSet[K]) ZRevData(rank int64) (key K, score float64) {
	return z.zData(rank, true)
}

// zData returns the id,score and extra data of an element which
// found by position in the rank.
// The parameter rank is the position, reverse says if in the descend rank.
func (z *ZSet[K]) zData(rank int64, reverse bool) (key K, score float64) {
	z.lock.RLock()
	defer z.lock.RUnlock()
	if rank < 0 || rank > z.zsl.length {
		return *new(K), 0
	}
	if reverse {
		rank = z.zsl.length - rank
	} else {
		rank++
	}
	n := z.zsl.zslGetElementByRank(uint64(rank))
	if n == nil {
		return *new(K), 0
	}
	score, ok := z.dict[n.id]
	if !ok {
		return *new(K), 0
	}
	return n.id, score
}

// ZRange implements ZRANGE
func (z *ZSet[K]) ZRange(start, end int64, f func(float64, K)) {
	z.snapshotRange(start, end, false, f)
}

// ZRevRange implements ZREVRANGE
func (z *ZSet[K]) ZRevRange(start, end int64, f func(float64, K)) {
	z.snapshotRange(start, end, true, f)
}

func (z *ZSet[K]) snapshotRange(start, end int64, reverse bool, f func(float64, K)) {
	scores := make([]float64, 0)
	keys := make([]K, 0)

	z.lock.RLock()
	z.commonRange(start, end, reverse, func(f float64, k K) {
		scores = append(scores, f)
		keys = append(keys, k)
	})
	z.lock.RUnlock()

	for i, score := range scores {
		f(score, keys[i])
	}
}

func (z *ZSet[K]) commonRange(start, end int64, reverse bool, f func(float64, K)) {
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

	var node *zNode[K]
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
