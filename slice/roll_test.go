package slice

import (
	"testing"
)

// TestRoll_SingleElement 修复前 rand.Int31n(0) panic。
func TestRoll_SingleElement(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Roll panicked on single element: %v", r)
		}
	}()
	got := Roll([]int{42})
	if got != 42 {
		t.Errorf("got %d, want 42", got)
	}
}

// TestRoll_CanPickLastElement 修复前 Int31n(l-1) 永不会返回最后一个索引。
func TestRoll_CanPickLastElement(t *testing.T) {
	nums := []int{1, 2}
	// 跑足够多次以几乎确定覆盖 index 1
	seen := map[int]bool{}
	for i := 0; i < 200; i++ {
		seen[Roll(nums)] = true
	}
	if !seen[1] || !seen[2] {
		t.Errorf("Roll should pick both elements over many samples, seen=%v", seen)
	}
}

func TestRoll_Empty(t *testing.T) {
	var zero int
	got := Roll([]int{})
	if got != zero {
		t.Errorf("Roll empty should return zero value, got %d", got)
	}
}
