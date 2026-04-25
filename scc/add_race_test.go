package scc

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// TestGO_WaitCountsAddedGoroutine 验证 Wait 会等到 GO 启动的 goroutine 完成。
// 修复前 Add(1) 在新 goroutine 内部,调度器不选它时 main 的 Wait 可能提前返回。
// 修复后 Add(1) 同步在 main 线程,Wait 必须观察到计数后才能返回。
// 注: 由于竞态本身依赖调度,此测试主要在 fix 生效后作为稳定正向断言,与 go vet 互补。
func TestGO_WaitCountsAddedGoroutine(t *testing.T) {
	s := New(context.TODO())
	ch := make(chan struct{})
	var done int32
	s.GO(func() {
		<-ch
		atomic.StoreInt32(&done, 1)
	})
	close(ch)
	// Cancel 抵消 New 时的初始 Add(1)
	s.Cancel()
	if err := s.Wait(2 * time.Second); err != nil {
		t.Fatalf("Wait timed out: %v", err)
	}
	if atomic.LoadInt32(&done) != 1 {
		t.Errorf("goroutine did not finish before Wait returned")
	}
}

func TestCGO_WaitCountsAddedGoroutine(t *testing.T) {
	s := New(context.TODO())
	ch := make(chan struct{})
	var done int32
	s.CGO(func(ctx context.Context) {
		<-ch
		atomic.StoreInt32(&done, 1)
	})
	close(ch)
	s.Cancel()
	if err := s.Wait(2 * time.Second); err != nil {
		t.Fatalf("Wait timed out: %v", err)
	}
	if atomic.LoadInt32(&done) != 1 {
		t.Errorf("CGO goroutine did not finish before Wait returned")
	}
}

func TestSGO_WaitCountsAddedGoroutine(t *testing.T) {
	s := New(context.TODO())
	ch := make(chan struct{})
	var done int32
	s.SGO(func(ctx context.Context) {
		<-ch
		atomic.StoreInt32(&done, 1)
	})
	close(ch)
	s.Cancel()
	if err := s.Wait(2 * time.Second); err != nil {
		t.Fatalf("Wait timed out: %v", err)
	}
	if atomic.LoadInt32(&done) != 1 {
		t.Errorf("SGO goroutine did not finish before Wait returned")
	}
}

// TestSGO_RecoverOnPanic 确认 SGO 的 recover 逻辑在重构后仍然工作。
func TestSGO_RecoverOnPanic(t *testing.T) {
	s := New(context.TODO())
	var caught int32
	s.Catch = func(err error) {
		atomic.StoreInt32(&caught, 1)
	}
	s.SGO(func(ctx context.Context) {
		panic("boom")
	})
	s.Cancel()
	if err := s.Wait(2 * time.Second); err != nil {
		t.Fatalf("Wait timed out: %v", err)
	}
	if atomic.LoadInt32(&caught) != 1 {
		t.Errorf("SGO did not catch panic via Catch hook")
	}
}
