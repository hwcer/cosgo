//Package zset is a port of t_zset.c in Redis
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
	"golang.org/x/exp/constraints"
	"math/rand"
)

const zSkipListMaxLevel = 32

type (
	// Key constraint
	//Key interface {
	//	constraints.Ordered
	//}
	zLevel[K constraints.Ordered] struct {
		forward *zNode[K]
		span    uint64
	}

	zNode[K constraints.Ordered] struct {
		id       K
		score    float64
		backward *zNode[K]
		level    []*zLevel[K]
	}

	skipList[K constraints.Ordered] struct {
		header *zNode[K]
		tail   *zNode[K]
		length int64
		level  int16
		order  int8 //积分相同时排序方式 -1 时间倒序，1 时间正序，  0 按key
	}

	zRangeSpec struct {
		min   float64
		max   float64
		minex int32
		maxex int32
	}
	zLexRangeSpec[K constraints.Ordered] struct {
		minKey K
		maxKey K
		minex  int
		maxex  int
	}
)

func zslCreateNode[K constraints.Ordered](level int16, score float64, id K) *zNode[K] {
	n := &zNode[K]{
		score: score,
		id:    id,
		level: make([]*zLevel[K], level),
	}
	for i := range n.level {
		n.level[i] = new(zLevel[K])
	}
	return n
}

func zslCreate[K constraints.Ordered](order ...int8) *skipList[K] {
	zsl := &skipList[K]{
		level:  1,
		header: zslCreateNode[K](zSkipListMaxLevel, 0, *new(K)),
	}
	if len(order) > 0 {
		zsl.order = order[0]
	}
	return zsl
}

const zSkipListP = 0.25 /* Skiplist P = 1/4 */

/* Returns a random zLevel for the new skiplist zNode we are going to create.
 * The return value of this function is between 1 and _ZSKIPLIST_MAXLEVEL
 * (both inclusive), with a powerlaw-alike distribution where higher
 * levels are less likely to be returned. */
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
func (zsl *skipList[K]) compare(k1, k2 K) bool {
	if zsl.order < 0 {
		return false
	} else if zsl.order > 0 {
		return true
	} else {
		return k1 < k2
	}
}

/* zslInsert a new zNode in the skiplist. Assumes the element does not already
 * exist (up to the caller to enforce that). The skiplist takes ownership
 * of the passed SDS string 'obj'. */
func (zsl *skipList[K]) zslInsert(score float64, id K) *zNode[K] {
	update := make([]*zNode[K], zSkipListMaxLevel)
	rank := make([]uint64, zSkipListMaxLevel)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		/* store rank that is crossed to reach the insert position */
		if i == zsl.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}
		if x.level[i] != nil {
			for x.level[i].forward != nil && (x.level[i].forward.score < score || (x.level[i].forward.score == score && zsl.compare(x.level[i].forward.id, id))) {
				rank[i] += x.level[i].span
				x = x.level[i].forward
			}
		}
		update[i] = x
	}
	/* we assume the element is not already inside, since we allow duplicated
	 * scores, reinserting the same element should never happen since the
	 * caller of zslInsert() should test in the hash table if the element is
	 * already inside or not. */
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

		/* update span covered by update[i] as x is inserted here */
		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	/* increment span for untouched levels */
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

/* Internal function used by zslDelete, zslDeleteByScore and zslDeleteByRank */
func (zsl *skipList[K]) zslDeleteNode(x *zNode[K], update []*zNode[K]) {
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

/* Delete an element with matching score/element from the skiplist.
 * The function returns 1 if the zNode was found and deleted, otherwise
 * 0 is returned.
 *
 * If 'zNode' is NULL the deleted zNode is freed by zslFreeNode(), otherwise
 * it is not freed (but just unlinked) and *zNode is set to the zNode pointer,
 * so that it is possible for the caller to reuse the zNode (including the
 * referenced SDS string at zNode->obj). */
func (zsl *skipList[K]) zslDelete(score float64, id K) int {
	update := make([]*zNode[K], zSkipListMaxLevel)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score ||
				(x.level[i].forward.score == score &&
					x.level[i].forward.id < id)) {
			x = x.level[i].forward
		}
		update[i] = x
	}
	/* We may have multiple elements with the same score, what we need
	 * is to find the element with both the right score and object. */
	x = x.level[0].forward
	if x != nil && score == x.score && x.id == id {
		zsl.zslDeleteNode(x, update)
		return 1
	}
	return 0 /* not found */
}

func zslValueGteMin(value float64, spec *zRangeSpec) bool {
	if spec.minex != 0 {
		return value > spec.min
	}
	return value >= spec.min
}

func zslValueLteMax(value float64, spec *zRangeSpec) bool {
	if spec.maxex != 0 {
		return value < spec.max
	}
	return value <= spec.max
}

/* Returns if there is a part of the zset is in range. */
func (zsl *skipList[K]) zslIsInRange(ran *zRangeSpec) bool {
	/* Test for ranges that will always be empty. */
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

/* Find the first zNode that is contained in the specified range.
 * Returns NULL when no element is contained in the range. */
func (zsl *skipList[K]) zslFirstInRange(ran *zRangeSpec) *zNode[K] {
	/* If everything is out of range, return early. */
	if !zsl.zslIsInRange(ran) {
		return nil
	}

	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		/* Go forward while *OUT* of range. */
		for x.level[i].forward != nil &&
			!zslValueGteMin(x.level[i].forward.score, ran) {
			x = x.level[i].forward
		}
	}
	/* This is an inner range, so the next zNode cannot be NULL. */
	x = x.level[0].forward
	//serverAssert(x != NULL);

	/* Check if score <= max. */
	if !zslValueLteMax(x.score, ran) {
		return nil
	}
	return x
}

/* Find the last zNode that is contained in the specified range.
 * Returns NULL when no element is contained in the range. */
func (zsl *skipList[K]) zslLastInRange(ran *zRangeSpec) *zNode[K] {

	/* If everything is out of range, return early. */
	if !zsl.zslIsInRange(ran) {
		return nil
	}
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		/* Go forward while *IN* range. */
		for x.level[i].forward != nil &&
			zslValueLteMax(x.level[i].forward.score, ran) {
			x = x.level[i].forward
		}
	}
	/* This is an inner range, so this zNode cannot be NULL. */
	//serverAssert(x != NULL);

	/* Check if score >= min. */
	if !zslValueGteMin(x.score, ran) {
		return nil
	}
	return x
}

/* Delete all the elements with score between min and max from the skiplist.
 * Min and max are inclusive, so a score >= min || score <= max is deleted.
 * Note that this function takes the reference to the hash table view of the
 * sorted set, in order to remove the elements from the hash table too. */
func (zsl *skipList[K]) zslDeleteRangeByScore(ran *zRangeSpec, dict map[K]float64) uint64 {
	removed := uint64(0)
	update := make([]*zNode[K], zSkipListMaxLevel)
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

	/* Current zNode is the last with score < or <= min. */
	x = x.level[0].forward

	/* Delete nodes while in range. */
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
		// Here is where x->obj is actually released.
		// And golang has GC, don't need to free manually anymore
		//zslFreeNode(x)
		removed++
		x = next
	}
	return removed
}

func (zsl *skipList[K]) zslDeleteRangeByLex(ran *zLexRangeSpec[K], dict map[K]float64) uint64 {
	removed := uint64(0)

	update := make([]*zNode[K], zSkipListMaxLevel)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && !zslLexValueGteMin(x.level[i].forward.id, ran) {
			x = x.level[i].forward
		}
		update[i] = x
	}

	/* Current zNode is the last with score < or <= min. */
	x = x.level[0].forward

	/* Delete nodes while in range. */
	for x != nil && zslLexValueLteMax(x.id, ran) {
		next := x.level[0].forward
		zsl.zslDeleteNode(x, update)
		delete(dict, x.id)
		removed++
		x = next
	}
	return removed
}

func zslLexValueGteMin[K constraints.Ordered](id K, spec *zLexRangeSpec[K]) bool {
	if spec.minex != 0 {
		return compareKey(id, spec.minKey) > 0
	}
	return compareKey(id, spec.minKey) >= 0
}

func compareKey[K constraints.Ordered](a, b K) int8 {
	if a == b {
		return 0
	} else if a > b {
		return 1
	}
	return -1
}

func zslLexValueLteMax[K constraints.Ordered](id K, spec *zLexRangeSpec[K]) bool {
	if spec.maxex != 0 {
		return compareKey(id, spec.maxKey) < 0
	}
	return compareKey(id, spec.maxKey) <= 0
}

/* Delete all the elements with rank between start and end from the skiplist.
 * Start and end are inclusive. Note that start and end need to be 1-based */
func (zsl *skipList[K]) zslDeleteRangeByRank(start, end uint64, dict map[K]float64) uint64 {
	update := make([]*zNode[K], zSkipListMaxLevel)
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

/* Find the rank for an element by both score and obj.
 * Returns 0 when the element cannot be found, rank otherwise.
 * Note that the rank is 1-based due to the span of zsl->header to the
 * first element. */
func (zsl *skipList[K]) zslGetRank(score float64, key K) int64 {
	rank := uint64(0)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score ||
				(x.level[i].forward.score == score &&
					x.level[i].forward.id <= key)) {
			rank += x.level[i].span
			x = x.level[i].forward
		}

		/* x might be equal to zsl->header, so test if obj is non-NULL */
		if x.id == key {
			return int64(rank)
		}
	}
	return 0
}

/* Finds an element by its rank. The rank argument needs to be 1-based. */
func (zsl *skipList[K]) zslGetElementByRank(rank uint64) *zNode[K] {
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
