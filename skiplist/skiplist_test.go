package skiplist

import (
	"fmt"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	sl := NewSkipList()
	rand.Seed(time.Now().UnixNano())

	// 模拟插入和排名查询
	for i := 0; i < 10; i++ {
		score := rand.Float64() * 100 // 生成0到100之间的随机分数
		sl.Insert(score, fmt.Sprintf("user%d", i))
		maxScore, _ := sl.GetMaxScore()
		rank := sl.GetRank(maxScore - 1)
	}
}
