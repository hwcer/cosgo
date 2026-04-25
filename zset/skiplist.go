/*
 * Copyright (c) 2009-2012, Salvatore Sanfilippo <antirez at gmail dot com>
 * Copyright (c) 2009-2012, Pieter Noordhuis <pcnoordhuis at gmail dot com>
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 *   * Redistributions of source code must retain the above copyright notice,
 *     this list of conditions and the following disclaimer.
 *   * Redistributions in binary form must reproduce the above copyright
 *     notice, this list of conditions and the following disclaimer in the
 *     documentation and/or other materials provided with the distribution.
 *   * Neither the name of Redis nor the names of its contributors may be used
 *     to endorse or promote products derived from this software without
 *     specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */

// 跳表（Skip List）实现
//
// 跳表是一种概率性数据结构，通过多层链表实现 O(log N) 的增删查改。
// 本实现支持降序（order < 0）和升序（order > 0）两种排列方式。
//
// 同分规则：相同分数的元素按插入顺序排列（FIFO，先到先得）。
// 这通过 insert 时使用 >= / <= 比较（跳过所有同分节点）来实现，
// 使新元素始终插入在已有同分元素之后。

package zset

import (
	"math/rand"
)

// zSkipListMaxLevel 跳表最大层数，32 层可支撑约 42 亿个元素
const zSkipListMaxLevel = 32

// zNodeInlineLevel 节点内联层数
// 层数 <= 此值的节点通过 combo 结构体实现单次堆分配（覆盖 99.6% 的节点）
const zNodeInlineLevel = 4

// zLevel 跳表节点在某一层的索引信息
type zLevel struct {
	forward *zNode // 同层下一个节点的指针
	span    int64  // 到 forward 节点的距离（跨越的节点数），用于计算排名
}

// zNode 跳表节点
type zNode struct {
	id       string  // 元素唯一标识（key）
	score    int64   // 元素分数，决定排序位置
	backward *zNode  // 第 0 层的前驱节点指针，用于反向遍历
	level    []zLevel // 各层索引（值类型切片），level[0] 是最底层
}

// zNodeCombo 将 zNode 和前 zNodeInlineLevel 层的 zLevel 打包在一起
// 使得绝大多数节点（层数 ≤ 4）只需一次堆分配
type zNodeCombo struct {
	node   zNode
	levels [zNodeInlineLevel]zLevel
}

// skipList 跳表
type skipList struct {
	header *zNode     // 头节点（哨兵节点，不存储实际数据）
	tail   *zNode     // 尾节点指针，指向最后一个实际节点
	length int64      // 当前节点总数（不含 header）
	level  int16      // 当前最高层数（从 1 开始）
	order  int8       // 排序方向：< 0 降序（高分在前），> 0 升序（低分在前）
	rng    *rand.Rand // 私有随机源，避免全局 rand 争用
}

// zslCreateNode 创建一个指定层数的跳表节点
// 层数 ≤ zNodeInlineLevel 时使用 combo 结构体，单次堆分配
// 层数 > zNodeInlineLevel 时回退到 zNode + make([]zLevel)，两次堆分配
func zslCreateNode(level int16, score int64, id string) *zNode {
	if level <= zNodeInlineLevel {
		c := new(zNodeCombo)
		c.node.score = score
		c.node.id = id
		c.node.level = c.levels[:level]
		return &c.node
	}
	return &zNode{
		score: score,
		id:    id,
		level: make([]zLevel, level),
	}
}

// zslCreate 创建并初始化一个空跳表
func zslCreate(order ...int8) *skipList {
	zsl := &skipList{
		level: 1,
		// header 层数为 zSkipListMaxLevel，走 make 分支（仅此一个节点）
		header: zslCreateNode(zSkipListMaxLevel, 0, ""),
		rng:    rand.New(rand.NewSource(rand.Int63())),
	}
	for i := range zsl.header.level {
		zsl.header.level[i].span = 1
	}
	if len(order) > 0 {
		zsl.order = order[0]
	}
	return zsl
}

// randomLevel 随机生成节点层数
// 使用私有随机源的单次 Uint64 调用，每 2 bit 决定一层（25% 概率提升）
func (zsl *skipList) randomLevel() int16 {
	bits := zsl.rng.Uint64()
	level := int16(1)
	for level < zSkipListMaxLevel && bits&3 == 0 {
		level++
		bits >>= 2
	}
	return level
}

// precedesScore 判断 score1 是否在排序中严格排在 score2 前面
func (zsl *skipList) precedesScore(score1, score2 int64) bool {
	if zsl.order < 0 {
		return score1 > score2
	}
	return score1 < score2
}

// precedesOrEqualScore 判断 score1 是否排在 score2 前面或与之相等
// 用于 insert：遍历时跳过所有同分节点，保证 FIFO
func (zsl *skipList) precedesOrEqualScore(score1, score2 int64) bool {
	if zsl.order < 0 {
		return score1 >= score2
	}
	return score1 <= score2
}

// rangeEntryScore 返回按分数范围查询时的入口分数
// 降序从 max 进入，升序从 min 进入
func (zsl *skipList) rangeEntryScore(min, max int64) int64 {
	if zsl.order < 0 {
		return max
	}
	return min
}

// rangeExitScore 返回按分数范围查询时的出口分数
// 降序到 min 结束，升序到 max 结束
func (zsl *skipList) rangeExitScore(min, max int64) int64 {
	if zsl.order < 0 {
		return min
	}
	return max
}

// zslInsert 向跳表中插入一个新节点
// 同分 FIFO：使用 precedesOrEqualScore 跳过所有同分节点，新元素插入在同分组末尾
func (zsl *skipList) zslInsert(score int64, id string) *zNode {
	var update [zSkipListMaxLevel]*zNode
	var rank [zSkipListMaxLevel]int64

	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		if i == zsl.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}
		for x.level[i].forward != nil &&
			zsl.precedesOrEqualScore(x.level[i].forward.score, score) {
			rank[i] += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}

	level := zsl.randomLevel()
	if level > zsl.level {
		for i := zsl.level; i < level; i++ {
			rank[i] = 0
			update[i] = zsl.header
			update[i].level[i].span = zsl.length
		}
		zsl.level = level
	}

	x = zslCreateNode(level, score, id)
	for i := int16(0); i < level; i++ {
		x.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = x
		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}
	for i := level; i < zsl.level; i++ {
		update[i].level[i].span++
	}

	if update[0] == zsl.header {
		x.backward = nil
	} else {
		x.backward = update[0]
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		zsl.tail = x
	}

	zsl.length++
	return x
}

// zslDeleteNode 从跳表中摘除节点 x，接收数组指针以保持栈分配
func (zsl *skipList) zslDeleteNode(x *zNode, update *[zSkipListMaxLevel]*zNode) {
	for i := int16(0); i < zsl.level; i++ {
		if update[i].level[i].forward == x {
			update[i].level[i].span += x.level[i].span - 1
			update[i].level[i].forward = x.level[i].forward
		} else {
			update[i].level[i].span--
		}
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x.backward
	} else {
		zsl.tail = x.backward
	}
	for zsl.level > 1 && zsl.header.level[zsl.level-1].forward == nil {
		zsl.level--
	}
	zsl.length--
}

// zslDelete 根据 (score, id) 删除指定节点
// 由于同分 FIFO，需在第 0 层扫描同分组定位目标，同时修正 update 数组
func (zsl *skipList) zslDelete(score int64, id string) bool {
	var update [zSkipListMaxLevel]*zNode
	x := zsl.header

	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			zsl.precedesScore(x.level[i].forward.score, score) {
			x = x.level[i].forward
		}
		update[i] = x
	}

	for x.level[0].forward != nil && x.level[0].forward.score == score && x.level[0].forward.id != id {
		x = x.level[0].forward
		for i := int16(0); i < int16(len(x.level)); i++ {
			update[i] = x
		}
	}

	x = x.level[0].forward
	if x != nil && x.score == score && x.id == id {
		zsl.zslDeleteNode(x, &update)
		return true
	}
	return false
}

// zslRank 获取元素 (score, key) 的 0-based 排名，找不到返回 -1
func (zsl *skipList) zslRank(score int64, key string) int64 {
	var rank int64
	x := zsl.header

	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			zsl.precedesScore(x.level[i].forward.score, score) {
			rank += x.level[i].span
			x = x.level[i].forward
		}
	}

	for x.level[0].forward != nil && x.level[0].forward.score == score {
		x = x.level[0].forward
		if x.id == key {
			return rank
		}
		rank++
	}
	return -1
}

// zslElement 获取指定 0-based 排名的节点，O(log n)
func (zsl *skipList) zslElement(rank int64) *zNode {
	target := rank + 1
	var traversed int64
	x := zsl.header

	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && (traversed+x.level[i].span) <= target {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		if traversed == target {
			return x
		}
	}
	return nil
}

// zslFirstInRange 返回分数范围 [min, max] 内第一个元素的 0-based 排名
// 不存在时返回 -1。O(log n)
func (zsl *skipList) zslFirstInRange(min, max int64) int64 {
	entry := zsl.rangeEntryScore(min, max)
	var rank int64
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && zsl.precedesScore(x.level[i].forward.score, entry) {
			rank += x.level[i].span
			x = x.level[i].forward
		}
	}
	x = x.level[0].forward
	if x == nil || x.score < min || x.score > max {
		return -1
	}
	return rank
}

// zslLastInRange 返回分数范围 [min, max] 内最后一个元素的 0-based 排名
// 不存在时返回 -1。O(log n)
func (zsl *skipList) zslLastInRange(min, max int64) int64 {
	exit := zsl.rangeExitScore(min, max)
	var rank int64
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && zsl.precedesOrEqualScore(x.level[i].forward.score, exit) {
			rank += x.level[i].span
			x = x.level[i].forward
		}
	}
	if x == zsl.header || x.score < min || x.score > max {
		return -1
	}
	return rank - 1
}

// zslCount 统计分数在 [min, max] 范围内的元素数量
// 通过两次 O(log n) 的 rank 查找做减法，无需线性遍历
func (zsl *skipList) zslCount(min, max int64) int64 {
	first := zsl.zslFirstInRange(min, max)
	if first < 0 {
		return 0
	}
	last := zsl.zslLastInRange(min, max)
	if last < 0 {
		return 0
	}
	return last - first + 1
}

// zslRange 返回排名在 [start, end]（0-based）范围内的节点列表，O(log n + k)
func (zsl *skipList) zslRange(start, end int64) []ZNode {
	if start > end {
		return nil
	}

	span := (end - start) + 1
	result := make([]ZNode, 0, span)

	x := zsl.header
	var traversed int64
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && (traversed+x.level[i].span) <= start {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
	}

	x = x.level[0].forward
	currentRank := start

	for x != nil && span > 0 {
		result = append(result, ZNode{
			Score: x.score,
			Key:   x.id,
			Rank:  currentRank,
		})
		x = x.level[0].forward
		currentRank++
		span--
	}

	return result
}

// zslRangeByScore 返回分数在 [min, max] 范围内的节点列表
// 先用 O(log n) 的 zslCount 预估容量一次性分配，避免 append 扩容
func (zsl *skipList) zslRangeByScore(min, max int64) []ZNode {
	count := zsl.zslCount(min, max)
	if count == 0 {
		return nil
	}

	result := make([]ZNode, 0, count)

	entry := zsl.rangeEntryScore(min, max)
	x := zsl.header
	var traversed int64
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && zsl.precedesScore(x.level[i].forward.score, entry) {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
	}

	x = x.level[0].forward
	currentRank := traversed

	for x != nil && x.score >= min && x.score <= max {
		result = append(result, ZNode{
			Score: x.score,
			Key:   x.id,
			Rank:  currentRank,
		})
		x = x.level[0].forward
		currentRank++
	}

	return result
}

// zslDeleteRangeByRank 删除排名在 [start, end]（0-based）范围内的所有节点
func (zsl *skipList) zslDeleteRangeByRank(start, end int64, dict map[string]int64) int64 {
	var update [zSkipListMaxLevel]*zNode
	var traversed, removed int64
	x := zsl.header

	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && (traversed+x.level[i].span) <= start {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}

	x = x.level[0].forward
	for x != nil && traversed <= end {
		next := x.level[0].forward
		zsl.zslDeleteNode(x, &update)
		delete(dict, x.id)
		removed++
		traversed++
		x = next
	}
	return removed
}

// zslDeleteRangeByScore 删除分数在 [min, max] 范围内的所有节点
func (zsl *skipList) zslDeleteRangeByScore(min, max int64, dict map[string]int64) int64 {
	var update [zSkipListMaxLevel]*zNode
	var removed int64

	entry := zsl.rangeEntryScore(min, max)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && zsl.precedesScore(x.level[i].forward.score, entry) {
			x = x.level[i].forward
		}
		update[i] = x
	}

	x = x.level[0].forward
	for x != nil && x.score >= min && x.score <= max {
		next := x.level[0].forward
		zsl.zslDeleteNode(x, &update)
		delete(dict, x.id)
		removed++
		x = next
	}

	return removed
}
