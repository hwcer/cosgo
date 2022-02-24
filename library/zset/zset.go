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

/*-----------------------------------------------------------------------------
 * Common sorted set API
 *----------------------------------------------------------------------------*/

// New creates a new SortedSet and return its pointer
func New() *SortedSet {
	s := &SortedSet{
		dict: make(map[int64]*obj),
		zsl:  zslCreate(),
	}
	return s
}

// Length returns counts of elements
func (z *SortedSet) Length() int64 {
	return z.zsl.length
}

// Set is used to add or update an element
func (z *SortedSet) Set(score float64, key int64, dat interface{}) {
	v, ok := z.dict[key]
	z.dict[key] = &obj{attachment: dat, key: key, score: score}
	if ok {
		/* destroy and re-insert when score changes. */
		if score != v.score {
			z.zsl.zslDelete(v.score, key)
			z.zsl.zslInsert(score, key)
		}
	} else {
		z.zsl.zslInsert(score, key)
	}
}

// IncrBy ..
func (z *SortedSet) IncrBy(score float64, key int64) (float64, interface{}) {
	v, ok := z.dict[key]
	if !ok {
		// use negative infinity ?
		return 0, nil
	}
	if score != 0 {
		z.zsl.zslDelete(v.score, key)
		v.score += score
		z.zsl.zslInsert(v.score, key)
	}
	return v.score, v.attachment
}

// Delete removes an element from the SortedSet
// by its key.
func (z *SortedSet) Delete(key int64) (ok bool) {
	v, ok := z.dict[key]
	if ok {
		z.zsl.zslDelete(v.score, key)
		delete(z.dict, key)
		return true
	}
	return false
}

// GetRank returns position,score and extra data of an element which
// found by the parameter key.
// The parameter reverse determines the rank is descent or ascend，
// true means descend and false means ascend.
func (z *SortedSet) GetRank(key int64, reverse bool) (rank int64, score float64, data interface{}) {
	v, ok := z.dict[key]
	if !ok {
		return -1, 0, nil
	}
	r := z.zsl.zslGetRank(v.score, key)
	if reverse {
		r = z.zsl.length - r
	} else {
		r--
	}
	return int64(r), v.score, v.attachment

}

// GetData returns data stored in the map by its key
func (z *SortedSet) GetData(key int64) (data interface{}, ok bool) {
	o, ok := z.dict[key]
	if !ok {
		return nil, false
	}
	return o.attachment, true
}

// GetDataByRank returns the id,score and extra data of an element which
// found by position in the rank.
// The parameter rank is the position, reverse says if in the descend rank.
func (z *SortedSet) GetDataByRank(rank int64, reverse bool) (key int64, score float64, data interface{}) {
	if rank < 0 || rank > z.zsl.length {
		return 0, 0, nil
	}
	if reverse {
		rank = z.zsl.length - rank
	} else {
		rank++
	}
	n := z.zsl.zslGetElementByRank(uint64(rank))
	if n == nil {
		return 0, 0, nil
	}
	dat, _ := z.dict[n.objID]
	if dat == nil {
		return 0, 0, nil
	}
	return dat.key, dat.score, dat.attachment
}

// Range implements ZRANGE
func (z *SortedSet) Range(start, end int64, f func(float64, int64, interface{})) {
	z.commonRange(start, end, false, f)
}

// RevRange implements ZREVRANGE
func (z *SortedSet) RevRange(start, end int64, f func(float64, int64, interface{})) {
	z.commonRange(start, end, true, f)
}

func (z *SortedSet) commonRange(start, end int64, reverse bool, f func(float64, int64, interface{})) {
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

	var node *skipListNode
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
		k := node.objID
		s := node.score
		f(s, k, z.dict[k].attachment)
		if reverse {
			node = node.backward
		} else {
			node = node.level[0].forward
		}
	}
}
