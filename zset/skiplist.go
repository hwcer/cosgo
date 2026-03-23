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
package zset

import (
	"math/rand"
)

const zSkipListMaxLevel = 32

// 基于 string 和 int64 的跳表实现
type zLevel struct {
	forward *zNode
	span    int64
}

type zNode struct {
	id       string
	score    int64
	backward *zNode
	level    []*zLevel
}

type skipList struct {
	header *zNode
	tail   *zNode
	length int64
	level  int16
	order  int8 // 排序方式，<0 倒序，>0 正序，0 按 key
}

func zslCreateNode(level int16, score int64, id string) *zNode {
	n := &zNode{
		score: score,
		id:    id,
		level: make([]*zLevel, level),
	}
	for i := range n.level {
		n.level[i] = new(zLevel)
	}
	return n
}

func zslCreate(order ...int8) *skipList {
	zsl := &skipList{
		level:  1,
		header: zslCreateNode(zSkipListMaxLevel, 0, ""),
	}
	// 初始化 header 节点的 span 为 1（表示到下一个节点的距离）
	for i := range zsl.header.level {
		zsl.header.level[i].span = 1
	}
	if len(order) > 0 {
		zsl.order = order[0]
	}
	return zsl
}

const zSkipListP = 0.25 /* Skiplist P = 1/4 */

func randomLevel() int16 {
	l := int16(1)
	for float32(rand.Int31()&0xFFFF) < (zSkipListP * 0xFFFF) {
		l++
	}
	if l < zSkipListMaxLevel {
		return l
	}
	return zSkipListMaxLevel
}

// shouldAdvance 判断在跳表遍历时是否应该继续前进
func (zsl *skipList) shouldAdvance(currentScore, targetScore int64) bool {
	if zsl.order < 0 {
		// 降序：高分在前，分数相同时新元素排在后面
		return currentScore >= targetScore
	} else {
		// 升序：低分在前，分数相同时新元素排在后面
		return currentScore <= targetScore
	}
}

func (zsl *skipList) zslInsert(score int64, id string) *zNode {
	update := make([]*zNode, zSkipListMaxLevel)
	rank := make([]int64, zSkipListMaxLevel)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		if i == zsl.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}
		if x.level[i] != nil {
			for x.level[i].forward != nil &&
				zsl.shouldAdvance(x.level[i].forward.score, score) {
				rank[i] += x.level[i].span
				x = x.level[i].forward
			}
		}
		update[i] = x
	}
	level := randomLevel()
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

func (zsl *skipList) zslDeleteNode(x *zNode, update []*zNode) {
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

func (zsl *skipList) zslDelete(score int64, id string) bool {
	update := make([]*zNode, zSkipListMaxLevel)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil {
			forward := x.level[i].forward
			// 在降序模式下，如果 forward.score > score，继续遍历
			// 在升序模式下，如果 forward.score < score，继续遍历
			// 如果分数相等，比较 id（字典序小的排在前面）
			if zsl.order < 0 {
				// 降序
				if forward.score > score || (forward.score == score && forward.id < id) {
					x = forward
					continue
				}
			} else {
				// 升序
				if forward.score < score || (forward.score == score && forward.id < id) {
					x = forward
					continue
				}
			}
			break
		}
		update[i] = x
	}
	x = x.level[0].forward
	if x != nil && score == x.score && x.id == id {
		zsl.zslDeleteNode(x, update)
		return true
	}
	return false
}

// zslRank 获取元素的排名（从0开始）
func (zsl *skipList) zslRank(score int64, key string) int64 {
	rank := int64(0)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			zsl.shouldAdvance(x.level[i].forward.score, score) {
			rank += x.level[i].span
			x = x.level[i].forward
		}
	}
	// 检查当前节点是否是目标节点
	// rank 是从 header 开始的距离，需要减 1 得到从 0 开始的排名
	if x != zsl.header && x.score == score && x.id == key {
		return rank - 1
	}
	return -1
}

// zslElement 获取指定排名的元素（从0开始）
func (zsl *skipList) zslElement(rank int64) *zNode {
	// 从 header 的 level[0] 开始遍历，找到第 rank 个元素
	x := zsl.header.level[0].forward
	traversed := int64(0)
	for x != nil && traversed < rank {
		x = x.level[0].forward
		traversed++
	}
	return x
}

// zslCount 统计分数在 min 和 max 之间的元素数量
func (zsl *skipList) zslCount(min, max int64) int64 {
	var count int64 = 0
	x := zsl.header.level[0].forward
	for x != nil {
		if x.score >= min && x.score <= max {
			count++
		} else if zsl.order < 0 {
			// 降序：当分数小于 min 时可以提前退出
			if x.score < min {
				break
			}
		} else {
			// 升序：当分数超过 max 时可以提前退出
			if x.score > max {
				break
			}
		}
		x = x.level[0].forward
	}
	return count
}

// zslRange 返回指定排名范围的元素节点
func (zsl *skipList) zslRange(start, end int64) []ZNode {
	if start > end {
		return nil
	}

	span := (end - start) + 1
	result := make([]ZNode, 0, span)

	// 获取起始节点
	x := zsl.header
	traversed := int64(0)
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

// zslRangeByScore 返回指定分数范围的元素节点
func (zsl *skipList) zslRangeByScore(min, max int64) []ZNode {
	var result []ZNode

	x := zsl.header.level[0].forward
	currentRank := int64(0)

	for x != nil {
		// 根据排序方式判断是否在范围内
		inRange := false
		if zsl.order < 0 {
			// 降序：分数从高到低
			inRange = x.score <= max && x.score >= min
		} else {
			// 升序：分数从低到高
			inRange = x.score >= min && x.score <= max
		}

		if inRange {
			result = append(result, ZNode{
				Score: x.score,
				Key:   x.id,
				Rank:  currentRank,
			})
		} else if zsl.order < 0 {
			// 降序：当分数小于 min 时可以提前退出
			if x.score < min {
				break
			}
		} else {
			// 升序：当分数超过 max 时可以提前退出
			if x.score > max {
				break
			}
		}

		x = x.level[0].forward
		currentRank++
	}

	return result
}

// zslDeleteRangeByRank 删除指定排名范围的元素（从0开始）
func (zsl *skipList) zslDeleteRangeByRank(start, end int64, dict map[string]int64) int64 {
	update := make([]*zNode, zSkipListMaxLevel)
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
		zsl.zslDeleteNode(x, update)
		delete(dict, x.id)
		removed++
		traversed++
		x = next
	}
	return removed
}

// zslDeleteRangeByScore 删除指定分数范围的元素
func (zsl *skipList) zslDeleteRangeByScore(min, max int64, dict map[string]int64) int64 {
	update := make([]*zNode, zSkipListMaxLevel)
	var removed int64

	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil {
			// 根据排序方式判断是否需要继续前进
			shouldContinue := false
			if zsl.order < 0 {
				// 降序：分数从高到低，找到第一个分数 <= max 的节点
				shouldContinue = x.level[i].forward.score > max
			} else {
				// 升序：分数从低到高，找到第一个分数 >= min 的节点
				shouldContinue = x.level[i].forward.score < min
			}
			if !shouldContinue {
				break
			}
			x = x.level[i].forward
		}
		update[i] = x
	}

	x = x.level[0].forward
	for x != nil {
		// 根据排序方式判断是否在范围内
		inRange := false
		if zsl.order < 0 {
			// 降序：分数从高到低
			inRange = x.score <= max && x.score >= min
		} else {
			// 升序：分数从低到高
			inRange = x.score >= min && x.score <= max
		}

		if !inRange {
			break
		}

		next := x.level[0].forward
		zsl.zslDeleteNode(x, update)
		delete(dict, x.id)
		removed++
		x = next
	}

	return removed
}
