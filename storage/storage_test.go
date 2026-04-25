package storage

import (
	"fmt"
	"sync"
	"testing"
)

func TestNewAndGet(t *testing.T) {
	s := New(100)

	setter := s.New("hello")
	if setter == nil {
		t.Fatal("New returned nil")
	}

	got, ok := s.Get(setter.Id())
	if !ok {
		t.Fatal("Get returned false for existing item")
	}
	if got.Id() != setter.Id() {
		t.Errorf("Id mismatch: got %s, want %s", got.Id(), setter.Id())
	}
}

func TestDelete(t *testing.T) {
	s := New(100)

	setter := s.New("data")
	id := setter.Id()

	deleted := s.Delete(id)
	if deleted == nil {
		t.Fatal("Delete returned nil for existing item")
	}
	if deleted.Id() != id {
		t.Errorf("Deleted Id mismatch: got %s, want %s", deleted.Id(), id)
	}

	_, ok := s.Get(id)
	if ok {
		t.Error("Get returned true after Delete")
	}
}

func TestDoubleDelete(t *testing.T) {
	s := New(100)

	setter := s.New("data")
	id := setter.Id()

	d1 := s.Delete(id)
	if d1 == nil {
		t.Fatal("First Delete returned nil")
	}

	d2 := s.Delete(id)
	if d2 != nil {
		t.Error("Second Delete should return nil")
	}
}

func TestSizeAndFree(t *testing.T) {
	s := New(10)

	if s.Size() != 0 {
		t.Errorf("Initial Size: got %d, want 0", s.Size())
	}
	if s.Free() != 10 {
		t.Errorf("Initial Free: got %d, want 10", s.Free())
	}

	var ids []string
	for i := 0; i < 5; i++ {
		setter := s.New(i)
		ids = append(ids, setter.Id())
	}

	if s.Size() != 5 {
		t.Errorf("After 5 inserts Size: got %d, want 5", s.Size())
	}
	if s.Free() != 5 {
		t.Errorf("After 5 inserts Free: got %d, want 5", s.Free())
	}

	s.Delete(ids[0])
	s.Delete(ids[1])

	if s.Size() != 3 {
		t.Errorf("After 2 deletes Size: got %d, want 3", s.Size())
	}
	if s.Free() != 7 {
		t.Errorf("After 2 deletes Free: got %d, want 7", s.Free())
	}
}

func TestBucketExpansion(t *testing.T) {
	s := New(5)

	for i := 0; i < 5; i++ {
		r := s.New(i)
		if r == nil {
			t.Fatalf("New returned nil at i=%d", i)
		}
	}

	if s.Size() != 5 {
		t.Errorf("Size: got %d, want 5", s.Size())
	}
	if s.Free() != 0 {
		t.Errorf("Free: got %d, want 0", s.Free())
	}

	// 第 6 个触发扩容
	r := s.New("overflow")
	if r == nil {
		t.Fatal("New returned nil after expansion")
	}
	if s.Size() != 6 {
		t.Errorf("Size after expansion: got %d, want 6", s.Size())
	}
	if s.Free() != 4 {
		t.Errorf("Free after expansion: got %d, want 4", s.Free())
	}

	// 验证新 ID 可以被 Get 到
	got, ok := s.Get(r.Id())
	if !ok || got.Id() != r.Id() {
		t.Error("Cannot Get item created during expansion")
	}
}

func TestSlotReuse(t *testing.T) {
	s := New(3)

	s1 := s.New("a")
	s2 := s.New("b")
	s3 := s.New("c")

	s.Delete(s2.Id())

	// 删除 s2 后空出一个槽位，新分配应复用该槽位（LIFO）
	s4 := s.New("d")
	if s4 == nil {
		t.Fatal("New returned nil after delete")
	}
	if s.Size() != 3 {
		t.Errorf("Size: got %d, want 3", s.Size())
	}

	// 所有 ID 应可正常获取
	for _, id := range []string{s1.Id(), s3.Id(), s4.Id()} {
		if _, ok := s.Get(id); !ok {
			t.Errorf("Get(%s) failed", id)
		}
	}
	// 旧 ID 不应被获取
	if _, ok := s.Get(s2.Id()); ok {
		t.Error("Get succeeded for deleted ID")
	}
}

func TestRange(t *testing.T) {
	s := New(10)

	ids := make(map[string]bool)
	for i := 0; i < 5; i++ {
		setter := s.New(i)
		ids[setter.Id()] = true
	}

	count := 0
	s.Range(func(setter Setter) bool {
		count++
		if !ids[setter.Id()] {
			t.Errorf("Range yielded unknown ID: %s", setter.Id())
		}
		return true
	})
	if count != 5 {
		t.Errorf("Range count: got %d, want 5", count)
	}
}

func TestRangeEarlyStop(t *testing.T) {
	s := New(10)
	for i := 0; i < 5; i++ {
		s.New(i)
	}

	count := 0
	s.Range(func(setter Setter) bool {
		count++
		return count < 3
	})
	if count != 3 {
		t.Errorf("Range early stop count: got %d, want 3", count)
	}
}

func TestInvalidGet(t *testing.T) {
	s := New(10)

	if _, ok := s.Get(""); ok {
		t.Error("Get('') should return false")
	}
	if _, ok := s.Get("invalid_id"); ok {
		t.Error("Get('invalid_id') should return false")
	}
}

func TestRemove(t *testing.T) {
	s := New(10)
	var ids []string
	for i := 0; i < 5; i++ {
		setter := s.New(i)
		ids = append(ids, setter.Id())
	}

	removed := s.Remove(ids[:3])
	if len(removed) != 3 {
		t.Errorf("Remove count: got %d, want 3", len(removed))
	}
	if s.Size() != 2 {
		t.Errorf("Size after Remove: got %d, want 2", s.Size())
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	s := New(100)

	const goroutines = 100
	const opsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// 写 goroutines: 循环 New + Delete
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				setter := s.New(fmt.Sprintf("data_%d_%d", id, j))
				if setter != nil {
					s.Delete(setter.Id())
				}
			}
		}(i)
	}

	// 读 goroutines: 循环 Range + Size
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				s.Range(func(setter Setter) bool {
					_ = setter.Id()
					return true
				})
				_ = s.Size()
				_ = s.Free()
			}
		}()
	}

	wg.Wait()

	// 所有对象都被 New + Delete，最终应为空
	if s.Size() != 0 {
		t.Errorf("Final Size: got %d, want 0", s.Size())
	}
}

func TestConcurrentExpansion(t *testing.T) {
	s := New(10) // 小桶，容易触发扩容

	const goroutines = 50
	const opsPerGoroutine = 20

	var wg sync.WaitGroup
	var mu sync.Mutex
	var allIds []string

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				setter := s.New(fmt.Sprintf("item_%d_%d", id, j))
				if setter != nil {
					mu.Lock()
					allIds = append(allIds, setter.Id())
					mu.Unlock()
				}
			}
		}(i)
	}
	wg.Wait()

	totalCreated := len(allIds)
	if s.Size() != totalCreated {
		t.Errorf("Size mismatch: got %d, want %d", s.Size(), totalCreated)
	}

	// 所有 ID 都应可 Get
	for _, id := range allIds {
		if _, ok := s.Get(id); !ok {
			t.Errorf("Get(%s) failed after concurrent creation", id)
		}
	}

	t.Logf("Created %d items across expansion, Size=%d, Free=%d", totalCreated, s.Size(), s.Free())
}

func TestFillAndDrain(t *testing.T) {
	s := New(100)

	var ids []string
	for i := 0; i < 100; i++ {
		setter := s.New(i)
		if setter == nil {
			t.Fatalf("New returned nil at i=%d", i)
		}
		ids = append(ids, setter.Id())
	}

	if s.Size() != 100 || s.Free() != 0 {
		t.Errorf("After fill: Size=%d Free=%d, want 100/0", s.Size(), s.Free())
	}

	for _, id := range ids {
		s.Delete(id)
	}

	if s.Size() != 0 || s.Free() != 100 {
		t.Errorf("After drain: Size=%d Free=%d, want 0/100", s.Size(), s.Free())
	}

	// 全部删除后应能重新分配
	setter := s.New("refill")
	if setter == nil {
		t.Fatal("New returned nil after draining")
	}
	if s.Size() != 1 {
		t.Errorf("After refill: Size=%d, want 1", s.Size())
	}
}

// ==================== Benchmark ====================

func BenchmarkBucketNew(b *testing.B) {
	bucket := NewBucket(0, b.N+1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bucket.New(i)
	}
}

func BenchmarkBucketGet(b *testing.B) {
	bucket := NewBucket(0, 10000)
	ids := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		s := bucket.New(i)
		ids[i] = s.Id()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bucket.Get(ids[i%10000])
	}
}

func BenchmarkStorageNewDelete(b *testing.B) {
	s := New(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		setter := s.New(i)
		s.Delete(setter.Id())
	}
}

func BenchmarkStorageGet(b *testing.B) {
	s := New(10000)
	ids := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		setter := s.New(i)
		ids[i] = setter.Id()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Get(ids[i%10000])
	}
}

func BenchmarkStorageSizeFree(b *testing.B) {
	s := New(10000)
	for i := 0; i < 5000; i++ {
		s.New(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Size()
		_ = s.Free()
	}
}

func BenchmarkParallelNewDelete(b *testing.B) {
	s := New(10000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			setter := s.New("x")
			if setter != nil {
				s.Delete(setter.Id())
			}
		}
	})
}
