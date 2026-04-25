package scc

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// TestTrigger_RunsBeforeCancel 注册的函数在 Cancel 时按序执行。
func TestTrigger_RunsBeforeCancel(t *testing.T) {
	s := New(context.TODO())
	var ran1, ran2 int32
	s.Trigger(func() { atomic.StoreInt32(&ran1, 1) })
	s.Trigger(func() { atomic.StoreInt32(&ran2, 1) })
	if !s.Cancel() {
		t.Fatal("Cancel returned false")
	}
	if atomic.LoadInt32(&ran1) != 1 {
		t.Errorf("f1 not run in Cancel")
	}
	if atomic.LoadInt32(&ran2) != 1 {
		t.Errorf("f2 not run in Cancel")
	}
}

// TestTrigger_SkipAfterCancel Cancel 之后再调用 Trigger 应当被丢弃,不执行不追加。
func TestTrigger_SkipAfterCancel(t *testing.T) {
	s := New(context.TODO())
	if !s.Cancel() {
		t.Fatal("Cancel returned false")
	}
	var ran int32
	s.Trigger(func() { atomic.StoreInt32(&ran, 1) })
	// 给调度一点时间,防止断言过早(其实 Trigger 不启动 goroutine,这里纯保险)
	time.Sleep(10 * time.Millisecond)
	if atomic.LoadInt32(&ran) != 0 {
		t.Errorf("Trigger after Cancel should be ignored, but fn ran")
	}
}

// TestCancel_Idempotent 第二次 Cancel 返回 false,不重复执行已注册函数。
func TestCancel_Idempotent(t *testing.T) {
	s := New(context.TODO())
	var n int32
	s.Trigger(func() { atomic.AddInt32(&n, 1) })
	if !s.Cancel() {
		t.Fatal("first Cancel returned false")
	}
	if s.Cancel() {
		t.Error("second Cancel should return false")
	}
	if atomic.LoadInt32(&n) != 1 {
		t.Errorf("trigger should run exactly once, got %d", n)
	}
}
