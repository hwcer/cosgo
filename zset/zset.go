/*
	https://www.epaperpress.com/sortsearch/download/skiplist.pdf
*/

package zset

const (
	SKIPLIST_MAXLEVEL    = 32   /* For 2^32 elements */
	SKIPLIST_Probability = 0.25 /* Skiplist probability = 1/4 */
)

type ZSet struct {
	records map[string]*Rank
}

/*
	ZSet public functions
*/

// New create a new sorted set
func New() *ZSet {
	return &ZSet{
		make(map[string]*Rank),
	}
}

func (z *ZSet) GetRank(key string, autoCreate ...bool) *Rank {
	rank := z.records[key]
	if rank == nil && len(autoCreate) > 0 && autoCreate[0] {
		rank = NewRank()
		z.records[key] = rank
	}
	return rank
}

func (z *ZSet) SetRank(key string, rank *Rank) {
	z.records[key] = rank
}

// ZAdd Adds the specified member with the specified score to the sorted set stored at key.
// Add an element into the sorted set with specific key / value / score.
// Time complexity of this method is : O(log(N))
func (z *ZSet) ZAdd(key string, score float64, member string, value interface{}) (val int) {
	rank := z.GetRank(key, true)
	return rank.ZAdd(score, member, value)
}

// ZScore returns the score of member in the sorted set at key.
func (z *ZSet) ZScore(key string, member string) (ok bool, score float64) {
	rank := z.GetRank(key)
	if rank == nil {
		return false, 0
	}
	node, exist := rank.dict[member]
	if !exist {
		return
	}
	return true, node.score
}

// ZCard returns the sorted set cardinality (number of elements) of the sorted set stored at key.
func (z *ZSet) ZCard(key string) int {
	rank := z.GetRank(key)
	if rank == nil {
		return 0
	}
	return len(rank.dict)
}

// ZRank returns the rank of member in the sorted set stored at key, with the scores ordered from low to high.
// The rank (or index) is 0-based, which means that the member with the lowest score has rank 0.
func (z *ZSet) ZRank(key, member string) int64 {
	rank := z.GetRank(key)
	if rank == nil {
		return -1
	}
	return rank.ZRank(member)
}

// ZRevRank returns the rank of member in the sorted set stored at key, with the scores ordered from high to low.
// The rank (or index) is 0-based, which means that the member with the highest score has rank 0.
func (z *ZSet) ZRevRank(key, member string) int64 {
	rank := z.GetRank(key)
	if rank == nil {
		return -1
	}
	return rank.ZRevRank(member)
}

// ZIncrBy increments the score of member in the sorted set stored at key by increment.
// If member does not exist in the sorted set, it is added with increment as its score (as if its previous score was 0.0).
// If key does not exist, a new sorted set with the specified member as its sole member is created.
func (z *ZSet) ZIncrBy(key string, increment float64, member string, value ...any) float64 {
	rank := z.GetRank(key, true)
	return rank.ZIncrBy(increment, member, value...)
}

// ZRem removes the specified members from the sorted set stored at key. Non existing members are ignored.
// An error is returned when key exists and does not hold a sorted set.
func (z *ZSet) ZRem(key, member string) bool {
	rank := z.GetRank(key)
	if rank == nil {
		return false
	}
	return rank.ZRem(member)
}

// ZScoreRange returns all the elements in the sorted set at key with a score between min and max (including elements with score equal to min or max).
// The elements are considered to be ordered from low to high scores.
func (z *ZSet) ZScoreRange(key string, min, max float64) (val []interface{}) {
	rank := z.GetRank(key)
	if rank == nil || min > max {
		return
	}

	return rank.ZScoreRange(min, max)
}

// ZRevScoreRange returns all the elements in the sorted set at key with a score between max and min (including elements with score equal to max or min).
// In contrary to the default ordering of sorted sets, for this command the elements are considered to be ordered from high to low scores.
func (z *ZSet) ZRevScoreRange(key string, max, min float64) (val []interface{}) {
	rank := z.GetRank(key)
	if rank == nil || max < min {
		return
	}
	return rank.ZRevScoreRange(max, min)
}

// ZKeyExists check if the key exists in Rank.
//func (z *ZSet) ZKeyExists(key string) bool {
//	rank := z.GetRank(key)
//	return rank != nil
//}
//
//// ZClear clear the key in Rank.
//func (z *ZSet) ZClear(key string) {
//	if z.ZKeyExists(key) {
//		delete(z.records, key)
//	}
//}

// ZRange returns the specified range of elements in the sorted set stored at <key>.
func (z *ZSet) ZRange(key string, start, stop int) []interface{} {
	rank := z.GetRank(key)
	if rank == nil {
		return nil
	}
	return rank.ZRange(start, stop)
}

// ZRangeWithScores returns the specified range of elements in the sorted set stored at <key>.
func (z *ZSet) ZRangeWithScores(key string, start, stop int) []interface{} {
	rank := z.GetRank(key)
	if rank == nil {
		return nil
	}
	return rank.ZRangeWithScores(start, stop)
}

// ZRevRange returns the specified range of elements in the sorted set stored at key.
// The elements are considered to be ordered from the highest to the lowest score.
// Descending lexicographical order is used for elements with equal score.
func (z *ZSet) ZRevRange(key string, start, stop int) []interface{} {
	rank := z.GetRank(key)
	if rank == nil {
		return nil
	}
	return rank.findRange(int64(start), int64(stop), true, false)
}

// ZRevRange returns the specified range of elements in the sorted set stored at key.
// The elements are considered to be ordered from the highest to the lowest score.
// Descending lexicographical order is used for elements with equal score.
func (z *ZSet) ZRevRangeWithScores(key string, start, stop int) []interface{} {
	rank := z.GetRank(key)
	if rank == nil {
		return nil
	}
	return rank.findRange(int64(start), int64(stop), true, true)
}

// ZGetByRank gets the member at key by rank, the rank is ordered from lowest to highest.
// The rank of lowest is 0 and so on.
func (z *ZSet) ZGetByRank(key string, rank int) []any {
	r := z.GetRank(key)
	if r == nil {
		return nil
	}
	return r.ZGetByRank(rank)
}

// ZRevGetByRank get the member at key by rank, the rank is ordered from highest to lowest.
// The rank of highest is 0 and so on.
func (z *ZSet) ZRevGetByRank(key string, rank int) (val []interface{}) {
	r := z.GetRank(key)
	if r == nil {
		return nil
	}
	return r.ZRevGetByRank(rank)
}

// ZPopMin get and remove the element with minimal score, nil if the set is empty
func (z *ZSet) ZPopMin(key string) (rec *Node) {
	r := z.GetRank(key)
	if r == nil {
		return nil
	}
	return r.ZPopMin()
}

// ZPopMax  get and remove the element with maximum score, nil if the set is empty
func (z *ZSet) ZPopMax(key string) (rec *Node) {
	r := z.GetRank(key)
	if r == nil {
		return nil
	}
	return r.ZPopMax()
}

/*
Returns all the elements in the sorted set at key with a score between min and max (including
elements with score equal to min or max). The elements are considered to be ordered from low to
high scores.

If options is nil, it searchs in interval [start, end] without any limit by default

https://github.com/wangjia184/sortedset/blob/af6d6d227aa79e2a64b899d995ce18aa0bef437c/sortedset.go#L283
*/
func (z *ZSet) ZRangeByScore(key string, start float64, end float64, options *ZRangeOptions) (nodes []*Node) {
	r := z.GetRank(key)
	if r == nil {
		return
	}
	return r.ZRangeByScore(start, end, options)
}

func (z *ZSet) Keys() []string {
	keys := make([]string, 0, len(z.records))
	for k := range z.records {
		keys = append(keys, k)
	}
	return keys
}
