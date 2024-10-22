package zset

import "math"

func NewRank() *Rank {
	return &Rank{
		dict: make(map[string]*Node),
		zsl:  NewSkipList(),
	}
}

type Rank struct {
	dict map[string]*Node
	zsl  *Skiplist
}

/*
	Rank node utility
*/

func (z *Rank) getNodeByRank(rank int64, reverse bool) (string, float64) {
	if rank < 0 || rank > z.zsl.length {
		return "", math.MinInt64
	}

	if reverse {
		rank = z.zsl.length - rank
	} else {
		rank++
	}

	n := z.zsl.getNodeByRank(uint64(rank))
	if n == nil {
		return "", math.MinInt64
	}

	node := z.dict[n.member]
	if node == nil {
		return "", math.MinInt64
	}

	return node.member, node.score

}

func (z *Rank) findRange(start, stop int64, reverse bool, withScores bool) (val []interface{}) {
	length := z.zsl.length

	if start < 0 {
		start += length
		if start < 0 {
			start = 0
		}
	}

	if stop < 0 {
		stop += length
	}

	if start > stop || start >= length {
		return
	}

	if stop >= length {
		stop = length - 1
	}
	span := (stop - start) + 1

	var node *Node
	if reverse {
		node = z.zsl.tail
		if start > 0 {
			node = z.zsl.getNodeByRank(uint64(length - start))
		}
	} else {
		node = z.zsl.head.level[0].forward
		if start > 0 {
			node = z.zsl.getNodeByRank(uint64(start + 1))
		}
	}

	for span > 0 {
		span--
		if withScores {
			val = append(val, node.member, node.score)
		} else {
			val = append(val, node.member)
		}
		if reverse {
			node = node.backward
		} else {
			node = node.level[0].forward
		}
	}

	return
}

// ZAdd Adds the specified member with the specified score to the sorted set stored at key.
// Add an element into the sorted set with specific key / value / score.
// Time complexity of this method is : O(log(N))
func (z *Rank) ZAdd(score float64, member string, value interface{}) (val int) {
	v, exist := z.dict[member]
	var node *Node
	if exist {
		val = 0
		// score changes, Delete and re-Insert
		if score != v.score {
			z.zsl.Delete(v.score, member)
			node = z.zsl.Insert(score, member, value)
		} else {
			// score does not change, update value
			v.value = value
		}
	} else {
		val = 1
		node = z.zsl.Insert(score, member, value)
	}

	if node != nil {
		z.dict[member] = node
	}

	return
}

// ZScore returns the score of member in the sorted set at key.
func (z *Rank) ZScore(member string) (ok bool, score float64) {
	node, exist := z.dict[member]
	if !exist {
		return
	}
	return true, node.score
}

// ZCard returns the sorted set cardinality (number of elements) of the sorted set stored at key.
func (z *Rank) ZCard() int {
	return len(z.dict)
}

// ZRank returns the rank of member in the sorted set stored at key, with the scores ordered from low to high.
// The rank (or index) is 0-based, which means that the member with the lowest score has rank 0.
func (z *Rank) ZRank(member string) int64 {
	v, exist := z.dict[member]
	if !exist {
		return -1
	}
	rank := z.zsl.Rank(v.score, member)
	rank--
	return rank
}

// ZRevRank returns the rank of member in the sorted set stored at key, with the scores ordered from high to low.
// The rank (or index) is 0-based, which means that the member with the highest score has rank 0.
func (z *Rank) ZRevRank(member string) int64 {
	v, exist := z.dict[member]
	if !exist {
		return -1
	}
	rank := z.zsl.Rank(v.score, member)
	return z.zsl.length - rank
}

// ZIncrBy increments the score of member in the sorted set stored at key by increment.
// If member does not exist in the sorted set, it is added with increment as its score (as if its previous score was 0.0).
// If key does not exist, a new sorted set with the specified member as its sole member is created.
func (z *Rank) ZIncrBy(increment float64, member string, value ...any) float64 {
	score := increment
	if node, ok := z.dict[member]; ok {
		score += node.score
	}
	var v any
	if len(value) > 0 {
		v = value[0]
	}
	z.ZAdd(increment, member, v)
	return increment
}

// ZRem removes the specified members from the sorted set stored at key. Non existing members are ignored.
// An error is returned when key exists and does not hold a sorted set.
func (z *Rank) ZRem(member string) bool {
	v, exist := z.dict[member]
	if exist {
		z.zsl.Delete(v.score, member)
		delete(z.dict, member)
		return true
	}
	return false
}

// ZScoreRange returns all the elements in the sorted set at key with a score between min and max (including elements with score equal to min or max).
// The elements are considered to be ordered from low to high scores.
func (z *Rank) ZScoreRange(min, max float64) (val []interface{}) {
	item := z.zsl
	minScore := item.head.level[0].forward.score
	if min < minScore {
		min = minScore
	}

	maxScore := item.tail.score
	if max > maxScore {
		max = maxScore
	}

	x := item.head
	for i := item.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && x.level[i].forward.score < min {
			x = x.level[i].forward
		}
	}

	x = x.level[0].forward
	for x != nil {
		if x.score > max {
			break
		}

		val = append(val, x.member, x.score)
		x = x.level[0].forward
	}

	return
}

// ZRevScoreRange returns all the elements in the sorted set at key with a score between max and min (including elements with score equal to max or min).
// In contrary to the default ordering of sorted sets, for this command the elements are considered to be ordered from high to low scores.
func (z *Rank) ZRevScoreRange(max, min float64) (val []interface{}) {
	item := z.zsl
	minScore := item.head.level[0].forward.score
	if min < minScore {
		min = minScore
	}

	maxScore := item.tail.score
	if max > maxScore {
		max = maxScore
	}

	x := item.head
	for i := item.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && x.level[i].forward.score <= max {
			x = x.level[i].forward
		}
	}
	for x != nil {
		if x.score < min {
			break
		}

		val = append(val, x.member, x.score)
		x = x.backward
	}
	return
}

// ZRange returns the specified range of elements in the sorted set stored at <key>.
func (z *Rank) ZRange(start, stop int) []interface{} {
	return z.findRange(int64(start), int64(stop), false, false)
}

// ZRangeWithScores returns the specified range of elements in the sorted set stored at <key>.
func (z *Rank) ZRangeWithScores(start, stop int) []interface{} {
	return z.findRange(int64(start), int64(stop), false, true)
}

// ZRevRange returns the specified range of elements in the sorted set stored at key.
// The elements are considered to be ordered from the highest to the lowest score.
// Descending lexicographical order is used for elements with equal score.
func (z *Rank) ZRevRange(start, stop int) []interface{} {
	return z.findRange(int64(start), int64(stop), true, false)
}

// ZRevRangeWithScores returns the specified range of elements in the sorted set stored at key.
// The elements are considered to be ordered from the highest to the lowest score.
// Descending lexicographical order is used for elements with equal score.
func (z *Rank) ZRevRangeWithScores(start, stop int) []interface{} {
	return z.findRange(int64(start), int64(stop), true, true)
}

// ZGetByRank gets the member at key by rank, the rank is ordered from lowest to highest.
// The rank of lowest is 0 and so on.
func (z *Rank) ZGetByRank(rank int) (val []interface{}) {
	member, score := z.getNodeByRank(int64(rank), false)
	val = append(val, member, score)
	return
}

// ZRevGetByRank get the member at key by rank, the rank is ordered from highest to lowest.
// The rank of highest is 0 and so on.
func (z *Rank) ZRevGetByRank(rank int) (val []interface{}) {
	member, score := z.getNodeByRank(int64(rank), true)
	val = append(val, member, score)
	return
}

// get and remove the element with minimal score, nil if the set is empty
func (z *Rank) ZPopMin() (rec *Node) {
	x := z.zsl.head.level[0].forward
	if x != nil {
		z.ZRem(x.member)
	}
	return x
}

// get and remove the element with maximum score, nil if the set is empty
func (z *Rank) ZPopMax() (rec *Node) {
	x := z.zsl.tail
	if x != nil {
		z.ZRem(x.member)
	}
	return x
}

type ZRangeOptions struct {
	Limit        int  // limit the max nodes to return
	ExcludeStart bool // exclude start value, so it search in interval (start, end] or (start, end)
	ExcludeEnd   bool // exclude end value, so it search in interval [start, end) or (start, end)
}

/*
Returns all the elements in the sorted set at key with a score between min and max (including
elements with score equal to min or max). The elements are considered to be ordered from low to
high scores.

If options is nil, it searchs in interval [start, end] without any limit by default

https://github.com/wangjia184/sortedset/blob/af6d6d227aa79e2a64b899d995ce18aa0bef437c/sortedset.go#L283
*/
func (z *Rank) ZRangeByScore(start float64, end float64, options *ZRangeOptions) (nodes []*Node) {
	zsl := z.zsl

	// prepare parameters
	var limit int = int((^uint(0)) >> 1)
	if options != nil && options.Limit > 0 {
		limit = options.Limit
	}

	excludeStart := options != nil && options.ExcludeStart
	excludeEnd := options != nil && options.ExcludeEnd
	reverse := start > end
	if reverse {
		start, end = end, start
		excludeStart, excludeEnd = excludeEnd, excludeStart
	}

	//determine if out of range
	if zsl.length == 0 {
		return nodes
	}

	if reverse { // search from end to start
		x := zsl.head

		if excludeEnd {
			for i := zsl.level - 1; i >= 0; i-- {
				for x.level[i].forward != nil &&
					x.level[i].forward.score < end {
					x = x.level[i].forward
				}
			}
		} else {
			for i := zsl.level - 1; i >= 0; i-- {
				for x.level[i].forward != nil &&
					x.level[i].forward.score <= end {
					x = x.level[i].forward
				}
			}
		}

		for x != nil && limit > 0 {
			if excludeStart {
				if x.score <= start {
					break
				}
			} else {
				if x.score < start {
					break
				}
			}

			next := x.backward

			nodes = append(nodes, x)
			limit--

			x = next
		}
	} else {
		// search from start to end
		x := zsl.head
		if excludeStart {
			for i := zsl.level - 1; i >= 0; i-- {
				for x.level[i].forward != nil &&
					x.level[i].forward.score <= start {
					x = x.level[i].forward
				}
			}
		} else {
			for i := zsl.level - 1; i >= 0; i-- {
				for x.level[i].forward != nil &&
					x.level[i].forward.score < start {
					x = x.level[i].forward
				}
			}
		}

		/* Current node is the last with score < or <= start. */
		x = x.level[0].forward

		for x != nil && limit > 0 {
			if excludeEnd {
				if x.score >= end {
					break
				}
			} else {
				if x.score > end {
					break
				}
			}

			next := x.level[0].forward

			nodes = append(nodes, x)
			limit--

			x = next
		}
	}

	return nodes
}
