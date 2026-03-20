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
	span    uint64
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

type zRangeSpec struct {
	min   int64
	max   int64
	minex int32
	maxex int32
}

type zLexRangeSpec struct {
	minKey string
	maxKey string
	minex  int
	maxex  int
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

func (zsl *skipList) compare(existingID, newID string) bool {
	if zsl.order > 0 {
		return true
	} else if zsl.order < 0 {
		return false
	} else {
		return existingID < newID
	}
}

// shouldAdvance 判断在跳表遍历时是否应该继续前进
func (zsl *skipList) shouldAdvance(currentScore, targetScore int64, currentID, targetID string) bool {
	if zsl.order < 0 {
		// 降序：高分在前
		return currentScore > targetScore || 
			(currentScore == targetScore && zsl.compare(currentID, targetID))
	} else {
		// 升序：低分在前
		return currentScore < targetScore || 
			(currentScore == targetScore && zsl.compare(currentID, targetID))
	}
}

func (zsl *skipList) zslInsert(score int64, id string) *zNode {
	update := make([]*zNode, zSkipListMaxLevel)
	rank := make([]uint64, zSkipListMaxLevel)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		if i == zsl.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}
		if x.level[i] != nil {
			for x.level[i].forward != nil && 
				zsl.shouldAdvance(x.level[i].forward.score, score, x.level[i].forward.id, id) {
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
			update[i].level[i].span = uint64(zsl.length)
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
		for x.level[i].forward != nil &&
			zsl.shouldAdvance(x.level[i].forward.score, score, x.level[i].forward.id, id) {
			x = x.level[i].forward
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

func zslValueGteMin(value int64, spec *zRangeSpec) bool {
	if spec.minex != 0 {
		return value > spec.min
	}
	return value >= spec.min
}

func zslValueLteMax(value int64, spec *zRangeSpec) bool {
	if spec.maxex != 0 {
		return value < spec.max
	}
	return value <= spec.max
}

func (zsl *skipList) zslIsInRange(ran *zRangeSpec) bool {
	if ran.min > ran.max ||
		(ran.min == ran.max && (ran.minex != 0 || ran.maxex != 0)) {
		return false
	}
	x := zsl.tail
	if x == nil || !zslValueGteMin(x.score, ran) {
		return false
	}
	x = zsl.header.level[0].forward
	if x == nil || !zslValueLteMax(x.score, ran) {
		return false
	}
	return true
}

func (zsl *skipList) zslFirstInRange(ran *zRangeSpec) *zNode {
	if !zsl.zslIsInRange(ran) {
		return nil
	}
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			!zslValueGteMin(x.level[i].forward.score, ran) {
			x = x.level[i].forward
		}
	}
	x = x.level[0].forward
	if !zslValueLteMax(x.score, ran) {
		return nil
	}
	return x
}

func (zsl *skipList) zslLastInRange(ran *zRangeSpec) *zNode {
	if !zsl.zslIsInRange(ran) {
		return nil
	}
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			zslValueLteMax(x.level[i].forward.score, ran) {
			x = x.level[i].forward
		}
	}
	if !zslValueGteMin(x.score, ran) {
		return nil
	}
	return x
}

func (zsl *skipList) zslDeleteRangeByScore(ran *zRangeSpec, dict map[string]int64) uint64 {
	removed := uint64(0)
	update := make([]*zNode, zSkipListMaxLevel)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil {
			var condition bool
			if ran.minex != 0 {
				condition = x.level[i].forward.score <= ran.min
			} else {
				condition = x.level[i].forward.score < ran.min
			}
			if !condition {
				break
			}
			x = x.level[i].forward
		}
		update[i] = x
	}
	x = x.level[0].forward
	for x != nil {
		var condition bool
		if ran.maxex != 0 {
			condition = x.score < ran.max
		} else {
			condition = x.score <= ran.max
		}
		if !condition {
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

func (zsl *skipList) zslDeleteRangeByLex(ran *zLexRangeSpec, dict map[string]int64) uint64 {
	removed := uint64(0)
	update := make([]*zNode, zSkipListMaxLevel)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && !zslLexValueGteMin(x.level[i].forward.id, ran) {
			x = x.level[i].forward
		}
		update[i] = x
	}
	x = x.level[0].forward
	for x != nil && zslLexValueLteMax(x.id, ran) {
		next := x.level[0].forward
		zsl.zslDeleteNode(x, update)
		delete(dict, x.id)
		removed++
		x = next
	}
	return removed
}

func zslLexValueGteMin(id string, spec *zLexRangeSpec) bool {
	if spec.minex != 0 {
		return compareKey(id, spec.minKey) > 0
	}
	return compareKey(id, spec.minKey) >= 0
}

func compareKey(a, b string) int8 {
	if a == b {
		return 0
	} else if a > b {
		return 1
	}
	return -1
}

func zslLexValueLteMax(id string, spec *zLexRangeSpec) bool {
	if spec.maxex != 0 {
		return compareKey(id, spec.maxKey) < 0
	}
	return compareKey(id, spec.maxKey) <= 0
}

func (zsl *skipList) zslDeleteRangeByRank(start, end uint64, dict map[string]int64) uint64 {
	update := make([]*zNode, zSkipListMaxLevel)
	var traversed, removed uint64
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && (traversed+x.level[i].span) < start {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}
	traversed++
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

func (zsl *skipList) zslGetRank(score int64, key string) int64 {
	rank := uint64(0)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			zsl.shouldAdvance(x.level[i].forward.score, score, x.level[i].forward.id, key) {
			rank += x.level[i].span
			x = x.level[i].forward
		}
		if x.level[i].forward != nil && x.level[i].forward.score == score && x.level[i].forward.id == key {
			return int64(rank + 1)
		}
	}
	return 0
}

func (zsl *skipList) zslGetElementByRank(rank uint64) *zNode {
	traversed := uint64(0)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && (traversed+x.level[i].span) <= rank {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		if traversed == rank {
			return x
		}
	}
	return nil
}

func (zsl *skipList) zslForceDeleteById(id string) int {
	if zsl.length == 0 {
		return 0
	}
	var deletedCount int = 0
	update := make([]*zNode, zSkipListMaxLevel)
	x := zsl.header.level[0].forward
	for x != nil {
		next := x.level[0].forward
		if x.id == id {
			for i := range update {
				update[i] = zsl.header
			}
			for i := zsl.level - 1; i >= 0; i-- {
				current := zsl.header
				for current.level[i].forward != nil && current.level[i].forward != x {
					current = current.level[i].forward
				}
				update[i] = current
			}
			zsl.zslDeleteNode(x, update)
			deletedCount++
		}
		x = next
	}
	return deletedCount
}

func (zsl *skipList) zslUpdateScore(oldScore, newScore int64, id string) bool {
	if zsl.length == 0 {
		return false
	}
	update := make([]*zNode, zSkipListMaxLevel)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && 
			zsl.shouldAdvance(x.level[i].forward.score, oldScore, x.level[i].forward.id, id) {
			x = x.level[i].forward
		}
		update[i] = x
	}
	if update[0] != zsl.header {
		x = update[0].level[0].forward
	} else {
		x = zsl.header.level[0].forward
	}
	for x != nil && x.score == oldScore {
		if x.id == id {
			needReposition := false
			if zsl.order < 0 {
				if newScore > oldScore {
					next := x.level[0].forward
					if next != nil && next.score < newScore {
						needReposition = true
					}
				} else if newScore < oldScore {
					prev := update[0]
					if prev != zsl.header && prev.score < newScore {
						needReposition = true
					}
				}
			} else {
				if newScore > oldScore {
					next := x.level[0].forward
					if next != nil && next.score > newScore {
						needReposition = true
					}
				} else if newScore < oldScore {
					prev := update[0]
					if prev != zsl.header && prev.score > newScore {
						needReposition = true
					}
				}
			}
			if !needReposition {
				x.score = newScore
				return true
			}
			zsl.zslDeleteNode(x, update)
			zsl.zslInsert(newScore, id)
			return false
		}
		x = x.level[0].forward
	}
	return false
}

func (zsl *skipList) zslUpdateOrInsert(score int64, id string, oldScore int64, exists bool) bool {
	if !exists {
		zsl.zslInsert(score, id)
		return false
	}
	if zsl.zslUpdateScore(oldScore, score, id) {
		return true
	}
	return false
}
