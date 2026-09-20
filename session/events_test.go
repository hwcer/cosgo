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

// 🔴 封板契约:启动完成后 On 只提示不注册——运行期注册会与 Emit 构成
// concurrent map fatal(不可恢复),拦下即可;误用不崩进程(Alert 留排查线索)。
// 契约:监听器仅允许在启动期注册(包 init 或启动钩子),运行期注册属编程错误
func TestOnAfterSealPanics(t *testing.T) {
	resetListeners()
	On(EventSessionRelease, func(v any) {})
	sealed := len(listeners[EventSessionRelease])
	if sealed == 0 {
		t.Fatal("前提:封板前注册应生效")
	}
	listenersSealed.Store(true)
	defer resetListeners()

	On(EventSessionRelease, func(v any) {}) //不得 panic

	if got := len(listeners[EventSessionRelease]); got != sealed {
		t.Fatalf("封板后的注册必须被忽略: 表内监听器 %d, 期望仍为 %d", got, sealed)
	}
}
