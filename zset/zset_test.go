package zset

import (
	"testing"
)

// 测试ZRange功能
func TestZRangeStringInt64(t *testing.T) {
	s := New(1)

	s.ZAdd(66, "1")
	s.ZAdd(66, "2")
	s.ZAdd(66, "3")
	s.ZAdd(100, "4")
	s.ZAdd(99, "5")
	s.ZAdd(33, "6")
	s.ZAdd(77, "7")

	rank, score := s.ZRank("2")
	t.Log("Key:", "2", "Rank:", rank, "Score:", score)

	var i int64
	s.ZRange(0, 10, func(v int64, k string) {
		i++
		t.Log("排序", "Key:", k, "Rank:", i, "Score:", v)
	})

	// 测试ZCount功能
	count := s.ZCount(60, 80)
	t.Log("ZCount(60, 80):", count)
	if count != 4 {
		t.Errorf("Expected count 4, got %d", count)
	}
}

// 测试ZRevRange功能
func TestZRevRangeStringInt64(t *testing.T) {
	s := New(-1)

	s.ZAdd(66, "1")
	s.ZAdd(66, "2")
	s.ZAdd(66, "3")
	s.ZAdd(100, "4")
	s.ZAdd(99, "5")
	s.ZAdd(33, "6")
	s.ZAdd(77, "7")

	rank, score := s.ZRevRank("2")
	t.Log("Key:", "2", "RevRank:", rank, "Score:", score)

	var i int64
	s.ZRevRange(0, 10, func(v int64, k string) {
		i++
		t.Log("逆序排序", "Key:", k, "Rank:", i, "Score:", v)
	})
}

// 测试ZIncr功能
func TestZIncrStringInt64(t *testing.T) {
	s := New()

	s.ZAdd(10, "user1")
	s.ZAdd(20, "user2")

	// 增加分数
	newScore := s.ZIncr(5, "user1")
	if newScore != 15 {
		t.Errorf("Expected new score 15, got %d", newScore)
	}
	t.Log("After ZIncr, user1 score:", newScore)

	// 测试ZCount在分数变化后的结果
	count := s.ZCount(10, 20)
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
	t.Log("ZCount(10, 20):", count)
}

// 测试ZRem功能
func TestZRemStringInt64(t *testing.T) {
	s := New()

	s.ZAdd(10, "user1")
	s.ZAdd(20, "user2")
	s.ZAdd(30, "user3")

	// 删除元素
	ok := s.ZRem("user2")
	if !ok {
		t.Error("Failed to remove user2")
	}
	t.Log("Remove user2 success:", ok)

	// 检查元素是否被删除
	score, exists := s.ZScore("user2")
	if exists {
		t.Errorf("user2 should not exist, but has score %d", score)
	}

	// 检查ZCount结果
	count := s.ZCount(0, 100)
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
	t.Log("After ZRem, ZCount(0, 100):", count)

	// 检查ZCard结果
	card := s.ZCard()
	if card != 2 {
		t.Errorf("Expected ZCard 2, got %d", card)
	}
	t.Log("After ZRem, ZCard:", card)
}

// 测试ZData和ZRevData功能
func TestZDataStringInt64(t *testing.T) {
	s := New()

	s.ZAdd(10, "user1")
	s.ZAdd(20, "user2")
	s.ZAdd(30, "user3")

	// 测试ZData
	key, score := s.ZData(0)
	if key != "user1" || score != 10 {
		t.Errorf("Expected (user1, 10), got (%s, %d)", key, score)
	}
	t.Log("ZData(0):", key, score)

	// 测试ZRevData
	key, score = s.ZRevData(0)
	if key != "user3" || score != 30 {
		t.Errorf("Expected (user3, 30), got (%s, %d)", key, score)
	}
	t.Log("ZRevData(0):", key, score)
}

// 测试ZScore功能
func TestZScoreStringInt64(t *testing.T) {
	s := New()

	s.ZAdd(100, "user1")
	s.ZAdd(200, "user2")

	// 测试存在的元素
	score, ok := s.ZScore("user1")
	if !ok || score != 100 {
		t.Errorf("Expected score 100, got %d (ok=%v)", score, ok)
	}
	t.Log("ZScore(user1):", score, "Exists:", ok)

	// 测试不存在的元素
	score, ok = s.ZScore("user3")
	if ok || score != 0 {
		t.Errorf("Expected score 0 and ok=false, got %d (ok=%v)", score, ok)
	}
	t.Log("ZScore(user3):", score, "Exists:", ok)
}
