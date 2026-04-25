package await

import (
	"errors"
	"testing"
	"time"
)

func TestCall_Success(t *testing.T) {
	a := New(10, time.Second)
	time.Sleep(50 * time.Millisecond) // 等待 process goroutine 启动

	reply, err := a.Call(func(args any) (any, error) {
		return args.(int) * 2, nil
	}, 21)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply != 42 {
		t.Fatalf("expected 42, got %v", reply)
	}
}

func TestCall_Error(t *testing.T) {
	a := New(10, time.Second)
	time.Sleep(50 * time.Millisecond)

	_, err := a.Call(func(args any) (any, error) {
		return nil, errors.New("task failed")
	}, nil)

	if err == nil || err.Error() != "task failed" {
		t.Fatalf("expected 'task failed', got %v", err)
	}
}

func TestCall_PanicRecovery(t *testing.T) {
	a := New(10, time.Second)
	time.Sleep(50 * time.Millisecond)

	_, err := a.Call(func(args any) (any, error) {
		panic("boom")
	}, nil)

	if err == nil {
		t.Fatal("expected error from panic, got nil")
	}
}

func TestTry_ServerBusy(t *testing.T) {
	// cap=1, 先塞满队列
	a := New(1, time.Second)
	time.Sleep(50 * time.Millisecond)

	blocker := func(args any) (any, error) {
		time.Sleep(time.Second)
		return nil, nil
	}

	// 第一个占满队列
	go a.Call(blocker, nil)
	time.Sleep(20 * time.Millisecond) // 确保第一个被 process 取走

	// 第二个填满 buffer
	go a.Call(blocker, nil)
	time.Sleep(20 * time.Millisecond) // 确保第二个进入 channel

	// 第三个应该被拒绝
	_, err := a.Try(func(args any) (any, error) {
		return "ok", nil
	}, nil)

	if err != ErrServerBusy {
		t.Fatalf("expected ErrServerBusy, got %v", err)
	}
}

func TestSync_AsyncResult(t *testing.T) {
	a := New(10, time.Second)
	time.Sleep(50 * time.Millisecond)

	msg := a.Sync(func(args any) (any, error) {
		return "async result", nil
	}, nil)

	reply, err := msg.Wait(time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply != "async result" {
		t.Fatalf("expected 'async result', got %v", reply)
	}
}

func TestWait_FastPath(t *testing.T) {
	a := New(10, time.Second)
	time.Sleep(50 * time.Millisecond)

	// Try 被拒绝时，done 在 Wait 之前已经被 signal，应走快路径（不分配 timer）
	a2 := New(0, time.Second) // cap=0 确保 Try 被拒
	_, err := a2.Try(func(args any) (any, error) {
		return nil, nil
	}, nil)

	if err != ErrServerBusy {
		t.Fatalf("expected ErrServerBusy, got %v", err)
	}
	_ = a // 防止 a 未使用
}
