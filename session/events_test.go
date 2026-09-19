package session

import (
	"sync/atomic"
	"testing"
)

// resetListeners 供每个测试隔离使用,恢复初始空表与未封板状态。
func resetListeners() {
	listeners = make(map[Event][]Listener)
	listenersSealed.Store(false)
}

// TestOnEmit_Basic 基本订阅与触发。
func TestOnEmit_Basic(t *testing.T) {
	resetListeners()
	var n atomic.Int32
	On(EventHeartbeat, func(v any) { n.Add(1) })
	Emit(EventHeartbeat, nil)
	Emit(EventHeartbeat, nil)
	if got := n.Load(); got != 2 {
		t.Errorf("got %d emits, want 2", got)
	}
}

// TestOnEmit_MultipleListeners 同一事件多个监听器按注册顺序全部触发。
func TestOnEmit_MultipleListeners(t *testing.T) {
	resetListeners()
	var calls []int
	for i := range 3 {
		id := i
		On(EventHeartbeat, func(v any) {
			calls = append(calls, id)
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

// 🔴 封板契约:启动完成后 On 必须 panic——运行期注册会与 Emit 构成
// concurrent map fatal(不可恢复),确定性 panic 优先于随机崩溃。
// 契约:监听器仅允许在启动期注册(包 init 或启动钩子),运行期注册属编程错误
func TestOnAfterSealPanics(t *testing.T) {
	resetListeners()
	listenersSealed.Store(true)
	defer resetListeners()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("封板后 On 应 panic")
		}
	}()
	On(EventSessionRelease, func(v any) {})
}
