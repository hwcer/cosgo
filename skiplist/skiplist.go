package skiplist

import (
	"container/list"
	"math/rand/v2"
)



//Node 定义跳表节点
type Node struct {
	score float64
	value interface{}
	back  *list.Element
	next  []*list.Element
}

// SkipList 跳表结构
type SkipList struct {
	header *list.List // 表头
	level  int        // 当前最大层数
}

//New 初始化跳表
func New() *SkipList {
	return &SkipList{
		header: list.New(),
		level:  0,
	}
}

//Insert 在跳表中插入一个元素
func (sl *SkipList) Insert(score float64, value interface{}) {
	// 在本地随机level
	level := 1
	for rand.Float64() < 0.5 {
		level++
		if level >= len(sl.header.Front().Value.(*Node).next {
			break
		}
	}

	// 创建新节点
	node := &Node{
		score: score,
		value: value,
		next:  make([]*list.Element, level),
	}

	// 在跳表中查找插入位置
	update := make([]*list.Element, level)
	current := sl.header
	for i := level - 1; i >= 0; i-- {
		for next := current.Next(); next != nil; next = current.Next() {
			if next.Value.(*Node).score > score {
				break
			}
			current = next
		}
		update[i] = current
	}

	// 将新节点链接到跳表
	for i := 0; i < level; i++ {
		node.next[i] = update[i].Next()
		update[i].Next() = sl.header.InsertAfter(node)
	}

	// 更新最大层数
	if level > sl.level {
		sl.level = level
	}
}

//GetMaxScore 获取跳表中的最高分数和对应的值
func (sl *SkipList) GetMaxScore() (float64, interface{}) {
	if sl.header.Len() == 0 {
		return 0, nil
	}
	return sl.header.Back().Value.(*Node).score, sl.header.Back().Value.(*Node).value
}

// GetRank 获取跳表排名在特定分数范围内的元素数量
func (sl *SkipList) GetRank(score float64) int {
	rank := 0
	current := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for next := current.Next(); next != nil; next = current.Next() {
			if next.Value.(*Node).score > score {
				break
			}
			rank++
			current = next
		}
		if current.Next() != nil {
			current = current.Next()
		}
	}
	return rank + 1
}
