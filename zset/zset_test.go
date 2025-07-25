package zset

import (
	"testing"
)

func TestZRange(t *testing.T) {
	s := New[int64](1)

	s.ZAdd(66, 1)
	s.ZAdd(66, 2)
	s.ZAdd(66, 3)
	s.ZAdd(100, 4)
	s.ZAdd(99, 5)
	s.ZAdd(33, 6)
	s.ZAdd(77, 7)

	rank, score := s.ZRank(2)
	t.Log("Key:", 2, "Rank:", rank, "Score:", score)

	var i int64
	s.ZRange(0, 10, func(v float64, k int64) {
		i++
		t.Log("排序", "Key:", k, "Rank:", i, "Score:", v)
	})

}

func TestZRevRange(t *testing.T) {
	s := New[int64](-1)

	s.ZAdd(66, 1)
	s.ZAdd(66, 2)
	s.ZAdd(66, 3)
	s.ZAdd(100, 4)
	s.ZAdd(99, 5)
	s.ZAdd(33, 6)
	s.ZAdd(77, 7)

	rank, score := s.ZRevRank(2)
	t.Log("Key:", 2, "Rank:", rank, "Score:", score)

	var i int64
	s.ZRevRange(0, 10, func(v float64, k int64) {
		i++
		t.Log("排序", "Key:", k, "Rank:", i, "Score:", v)
	})

}
