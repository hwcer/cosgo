package session

import (
	"sync"
	"sync/atomic"
	"testing"
)

// resetListeners 供每个测试隔离使用,恢复初始空 map。
func resetListeners() {
	empty := map[Event][]Listener{}
	listenersV.Store(&empty)
}

// TestOnEmit_Basic 基本订阅与触发。
func TestOnEmit_Basic(t *testing.T) {
	resetListeners()
	var n int32
	On(EventHeartbeat, func(v any) { atomic.AddInt32(&n, 1) })
	Emit(EventHeartbeat, nil)
	Emit(EventHeartbeat, nil)
	if got := atomic.LoadInt32(&n); got != 2 {
		t.Errorf("got %d emits, want 2", got)
	}
}

// TestOnEmit_MultipleListeners 同一事件多个监听器按注册顺序全部触发。
func TestOnEmit_MultipleListeners(t *testing.T) {
	resetListeners()
	var calls []int
	var mu sync.Mutex
	for i := 0; i < 3; i++ {
		id := i
		On(EventHeartbeat, func(v any) {
			mu.Lock()
			calls = append(calls, id)
			mu.Unlock()
		})
	}
	Emit(EventHeartbeat, nil)
	if len(calls) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(calls))
	}
	for i, id := range calls {
		if id != i {
			t.Errorf("order mismatch at %d: got %d", i, id)
		}
	}
}

// TestOnEmit_ConcurrentOnAndEmit 大量并发 On + Emit 不触发 race 或 panic,
// 监听器最终计数等于 Emit 发生时已发布的订阅数之和(允许有"晚来的 On 不被早期 Emit 看到")。
func TestOnEmit_ConcurrentOnAndEmit(t *testing.T) {
	resetListeners()
	var emits int32
	var wg sync.WaitGroup

	// 并发写者: 注册一堆监听器
	const writers = 20
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			On(EventHeartbeat, func(v any) {
				atomic.AddInt32(&emits, 1)
			})
		}()
	}
	// 并发读者: 在注册过程中触发事件
	const readers = 10
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				Emit(EventHeartbeat, nil)
			}
		}()
	}
	wg.Wait()

	// 注册完成后,再发一次,应看到全部 writers 个监听器
	atomic.StoreInt32(&emits, 0)
	Emit(EventHeartbeat, nil)
	if got := atomic.LoadInt32(&emits); got != writers {
		t.Errorf("after all Ons, Emit should hit %d listeners, got %d", writers, got)
	}
}

// TestOn_DoesNotMutateOldSnapshot 验证 copy-on-write: 持有旧快照的读者看到的数据不变。
func TestOn_DoesNotMutateOldSnapshot(t *testing.T) {
	resetListeners()
	On(EventHeartbeat, func(v any) {})
	oldSnap := listenersV.Load()
	oldSlice := (*oldSnap)[EventHeartbeat]
	oldLen := len(oldSlice)
	// 追加新监听器
	On(EventHeartbeat, func(v any) {})
	// 旧 slice 不应被影响
	if len((*oldSnap)[EventHeartbeat]) != oldLen {
		t.Errorf("old snapshot mutated: len was %d, now %d", oldLen, len((*oldSnap)[EventHeartbeat]))
	}
}
