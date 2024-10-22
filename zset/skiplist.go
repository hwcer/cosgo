package zset

import "math/rand"

type Level struct {
	forward *Node
	span    uint64
}

// (hop table)
type Node struct {
	member   string
	value    interface{}
	score    float64
	backward *Node
	level    []*Level
}

// Skiplist Node in skip list (jump table)
type Skiplist struct {
	head   *Node
	tail   *Node
	length int64
	level  int
}

// Returns a random level for the new skiplist node we are going to create.
// The return value of this function is between 1 and SKIPLIST_MAXLEVEL
// (both inclusive), with a powerlaw-alike distribution where higher
// levels are less likely to be returned.
func randomLevel() int {
	level := 1
	for float64(rand.Int31()&0xFFFF) < float64(SKIPLIST_Probability*0xFFFF) {
		level += 1
	}
	if level < SKIPLIST_MAXLEVEL {
		return level
	}

	return SKIPLIST_MAXLEVEL
}

func createNode(level int, score float64, member string, value interface{}) *Node {
	node := &Node{
		score:  score,
		member: member,
		value:  value,
		level:  make([]*Level, level),
	}

	for i := range node.level {
		node.level[i] = new(Level)
	}

	return node
}

func NewSkipList() *Skiplist {
	return &Skiplist{
		level: 1,
		head:  createNode(SKIPLIST_MAXLEVEL, 0, "", nil),
	}
}

/*
Insert a new node in the skiplist. Assumes the element does not already
exist (up to the caller to enforce that). The skiplist takes ownership
of the passed member string.
*/
func (z *Skiplist) Insert(score float64, member string, value interface{}) *Node {
	/*

		https://www.youtube.com/watch?v=UGaOXaXAM5M
		https://www.youtube.com/watch?v=NDGpsfwAaqo

		The update array stores previous pointers for each level, new node
		will be added after them. rank array stores the rank value of each skiplist node.

		Steps:

		generate update and rank array
		create a new node with random level
		Insert new node according to update and rank info
		update other necessary infos, such as span, backward pointer, length.
	*/

	updates := make([]*Node, SKIPLIST_MAXLEVEL)
	rank := make([]uint64, SKIPLIST_MAXLEVEL)

	x := z.head
	for i := z.level - 1; i >= 0; i-- {
		/* store rank that is crossed to reach the Insert position */
		if i == z.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		if x.level[i] != nil {
			for x.level[i].forward != nil &&
				(x.level[i].forward.score < score ||
					(x.level[i].forward.score == score && // score is the same but the key is different
						x.level[i].forward.member < member)) {

				rank[i] += x.level[i].span
				x = x.level[i].forward
			}
		}
		updates[i] = x
	}

	/* we assume the key is not already inside, since we allow duplicated
	 * scores, and the re-insertion of score and redis object should never
	 * happen since the caller of Insert() should test in the hash table
	 * if the element is already inside or not. */
	level := randomLevel()
	if level > z.level { // add a new level
		for i := z.level; i < level; i++ {
			rank[i] = 0
			updates[i] = z.head
			updates[i].level[i].span = uint64(z.length)
		}
		z.level = level
	}

	x = createNode(level, score, member, value)
	for i := 0; i < level; i++ {
		x.level[i].forward = updates[i].level[i].forward
		updates[i].level[i].forward = x

		/* update span covered by update[i] as x is inserted here */
		x.level[i].span = updates[i].level[i].span - (rank[0] - rank[i])
		updates[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	/* increment span for untouched levels */
	for i := level; i < z.level; i++ {
		updates[i].level[i].span++
	}

	if updates[0] == z.head {
		x.backward = nil
	} else {
		x.backward = updates[0]
	}

	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		z.tail = x
	}

	z.length++
	return x
}

// Delete an element with matching score/key from the skiplist.
func (z *Skiplist) Delete(score float64, member string) {
	update := make([]*Node, SKIPLIST_MAXLEVEL)

	x := z.head
	for i := z.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score ||
				(x.level[i].forward.score == score && x.level[i].forward.member < member)) {
			x = x.level[i].forward
		}
		update[i] = x
	}

	/* We may have multiple elements with the same score, what we need
	 * is to find the element with both the right score and object. */
	x = x.level[0].forward
	if x != nil && score == x.score && x.member == member {
		z.deleteNode(x, update)
		return
	}
}

/* Internal function used by Delete, DeleteByScore and DeleteByRank */
func (z *Skiplist) deleteNode(x *Node, updates []*Node) {
	for i := 0; i < z.level; i++ {
		if updates[i].level[i].forward == x {
			updates[i].level[i].span += x.level[i].span - 1
			updates[i].level[i].forward = x.level[i].forward
		} else {
			updates[i].level[i].span--
		}
	}

	if x.level[0].forward != nil {
		x.level[0].forward.backward = x.backward
	} else {
		z.tail = x.backward
	}

	for z.level > 1 && z.head.level[z.level-1].forward == nil {
		z.level--
	}

	z.length--
}

// Rank Find the rank of the node specified by key
// Note that the rank is 0-based integer. Rank 0 means the first node
func (z *Skiplist) Rank(score float64, member string) int64 {
	var rank uint64 = 0
	x := z.head
	for i := z.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score ||
				(x.level[i].forward.score == score &&
					x.level[i].forward.member <= member)) {
			rank += x.level[i].span
			x = x.level[i].forward
		}

		if x.member == member {
			return int64(rank)
		}
	}

	return 0
}

func (z *Skiplist) getNodeByRank(rank uint64) *Node {
	var traversed uint64 = 0

	x := z.head
	for i := z.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(traversed+x.level[i].span) <= rank {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		if traversed == rank {
			return x
		}
	}

	return nil
}
