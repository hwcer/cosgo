package zset

const (
	// cleanupBufferSize 清理缓冲大小
	// 当超出人数限制超过此值时才触发清理
	cleanupBufferSize = 100
)

// ZSet 是针对 string 类型 key 和 int64 类型分数的有序集合实现（无锁版本）
type ZSet struct {
	dict    map[string]int64
	zsl     *skipList
	maxSize int32       // 最大人数限制，0=不限制
	guard   guardKeeper // 守门员（最后一名分数）
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
func (z *ZSet) canEnter(score int64, key string, exists bool) bool {
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

	// 根据排序方向判断
	if z.zsl.order < 0 {
		// 降序：需要比守门员分数高才能进入
		return score > z.guard.score
	} else {
		// 升序：需要比守门员分数低才能进入
		return score < z.guard.score
	}
}

// zadd 内部方法，用于添加或更新元素，返回最终的分数
func (z *ZSet) zadd(score int64, key string) int64 {
	// 首先检查 key 是否已存在
	oldScore, exists := z.dict[key]

	// 如果 key 已存在且分数相同，则不需要修改
	if exists && score == oldScore {
		return score
	}

	// 统一检查是否能进入排行榜（只调用一次）
	canEnter := z.canEnter(score, key, exists)

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

// ZAdd 用于添加或更新元素（无锁版本，需要外部同步）
func (z *ZSet) ZAdd(score int64, key string) {
	z.zadd(score, key)
}

// ZIncr 有序集合中对指定成员的分数加上增量 score（无锁版本）
func (z *ZSet) ZIncr(score int64, key string) int64 {
	if score == 0 {
		// 如果增量为 0，直接返回当前分数
		if currentScore, ok := z.dict[key]; ok {
			return currentScore
		}
		return 0
	}
	newScore := z.dict[key] + score
	return z.zadd(newScore, key)
}

// ZRem 通过 key 从 ZSet 中删除元素（无锁版本）
func (z *ZSet) ZRem(key string) bool {
	score, ok := z.dict[key]
	if !ok {
		return false
	}

	// 先验证跳表删除是否成功
	deleted := z.zsl.zslDelete(score, key)
	if deleted {
		delete(z.dict, key)

		// 如果删除的是守门员，需要重新计算
		if z.guard.valid && key == z.guard.key {
			z.updateGuard()
		}

		return true
	}

	// 跳表删除失败，可能是在守门员外的元素
	// 只删除 dict
	delete(z.dict, key)
	return true
}

// ZRank 返回元素的排名（从 0 开始）和分数
func (z *ZSet) ZRank(key string) (rank int64, score int64) {
	score, ok := z.dict[key]
	if !ok {
		return -1, 0
	}
	r := z.zsl.zslGetRank(score, key)
	return r - 1, score
}

// ZScore 通过 key 获取分数（无锁版本）
func (z *ZSet) ZScore(key string) (score int64, ok bool) {
	score, ok = z.dict[key]
	return score, ok
}

// ZData 返回指定排名的元素的 key 和分数
func (z *ZSet) ZData(rank int64) (key string, score int64) {
	if rank < 0 || rank >= z.zsl.length {
		return "", 0
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

// ZRange 按照跳表顺序遍历指定范围的元素
func (z *ZSet) ZRange(start, end int64, f func(int64, string)) {
	z.commonRange(start, end, f)
}

// commonRange 内部方法，实现范围查询的通用逻辑
func (z *ZSet) commonRange(start, end int64, f func(int64, string)) {
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

	// 从头部开始向后遍历
	node := z.zsl.header.level[0].forward
	if start > 0 {
		node = z.zsl.zslGetElementByRank(uint64(start + 1))
	}

	for span > 0 {
		span--
		k := node.id
		s := node.score
		f(s, k)
		node = node.level[0].forward
	}
}

// ZCount 返回有序集合中分数在 min 和 max 之间的元素数量
func (z *ZSet) ZCount(min, max int64) int64 {
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
			// 因为是有序的，所以当分数超过 max 时可以提前退出
			break
		}
	}
	return count
}

// GetGuardScore 获取守门员分数（最后一名的分数）
func (z *ZSet) GetGuardScore() (int64, bool) {
	if !z.guard.valid {
		return 0, false
	}

	return z.guard.score, true
}

// IsFull 检查是否已满员
func (z *ZSet) IsFull() bool {
	if z.maxSize <= 0 {
		return false
	}

	return int64(len(z.dict)) >= int64(z.maxSize)
}

// SetMaxSize 设置最大人数限制
func (z *ZSet) SetMaxSize(maxSize int32) {
	z.maxSize = maxSize
	// 重新计算守门员
	z.updateGuard()
	// 可能需要裁剪
	z.tryTrimExcess()
}

// GetMaxSize 获取最大人数限制
func (z *ZSet) GetMaxSize() int32 {
	return z.maxSize
}

// ZCard 返回元素数量
func (z *ZSet) ZCard() int64 {
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
	key, score := z.ZData(rank)
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

	// 触发清理
	z.trimExcess()
}

// isAfterGuard 判断元素是否在守门员之后
func (z *ZSet) isAfterGuard(score int64) bool {
	if !z.guard.valid {
		return false
	}

	if z.zsl.order < 0 {
		// 降序：分数小于等于守门员分数的元素在守门员之后
		return score <= z.guard.score
	} else {
		// 升序：分数大于等于守门员分数的元素在守门员之后
		return score >= z.guard.score
	}
}

// trimExcess 清理超出阈值的元素
func (z *ZSet) trimExcess() {
	if z.maxSize <= 0 {
		return
	}

	// 1. 清理跳表中超出 maxSize 的元素
	if z.zsl.length > int64(z.maxSize) {
		// 删除从 maxSize+1 到末尾的元素
		z.deleteRangeByRank(int64(z.maxSize)+1, z.zsl.length)
	}

	// 2. 只有当守门员有效时，才清理 dict 中不在跳表的元素
	if z.guard.valid {
		for key, score := range z.dict {
			if z.isAfterGuard(score) {
				delete(z.dict, key)
			}
		}
	}
}

// deleteRangeByRank 删除指定排名范围的元素
func (z *ZSet) deleteRangeByRank(start, end int64) {
	if start > end || start == 0 {
		return
	}

	// 直接使用跳表的批量删除方法（O(K + log N)）
	z.zsl.zslDeleteRangeByRank(uint64(start), uint64(end), z.dict)
}
