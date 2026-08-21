package cosgo

import (
	"errors"
	"strings"
	"testing"
)

func resetEvents() { events = make(map[EventType][]EventFunc) }

// TestEmitPanicKeepsStack panic 必须带出【原始堆栈】与【是哪个监听器】
//
// 🔴 这是本文件存在的理由。事件监听器由各模块在自己的 init 里 On() 进来，
// 顺序对使用者不可见；一旦某个监听器 nil 解引用，旧实现只给出
// "cosgo emit[1] error: invalid memory address" + cosgo.Start→main.main 两帧，
// 业务层帧全被 recover 回卷掉了，等于什么线索都没有。
func TestEmitPanicKeepsStack(t *testing.T) {
	defer resetEvents()
	resetEvents()

	On(EventTypLoaded, func() error { return nil })
	On(EventTypLoaded, func() error {
		var p *struct{ X int }
		_ = p.X // 故意 nil 解引用
		return nil
	})

	err := emit(EventTypLoaded, true)
	if err == nil {
		t.Fatal("panic 必须转成 error")
	}
	s := err.Error()
	for _, want := range []string{
		"EventTypLoaded",          // 事件名，不是裸数字
		"第 1 个",                   // 哪一个监听器（0 基）
		"TestEmitPanicKeepsStack", // 原始堆栈里必须有 panic 现场所在的函数
	} {
		if !strings.Contains(s, want) {
			t.Errorf("错误信息缺少 %q:\n%s", want, s)
		}
	}
}

// TestEmitErrorNamesHandler 返回 error 的路径同样要指名道姓
func TestEmitErrorNamesHandler(t *testing.T) {
	defer resetEvents()
	resetEvents()

	sentinel := errors.New("boom")
	On(EventTypBegin, func() error { return sentinel })

	err := emit(EventTypBegin, true)
	if err == nil {
		t.Fatal("应返回错误")
	}
	//必须可用 errors.Is 追到原始错误——包装不能把它埋掉
	if !errors.Is(err, sentinel) {
		t.Errorf("包装后应仍能 errors.Is 到原错误: %v", err)
	}
	if !strings.Contains(err.Error(), "EventTypBegin") {
		t.Errorf("错误信息应带事件名: %v", err)
	}
}

// TestEmitBreakOnError 短路与 best-effort 两种语义
//
// panic 与「返回 error」一视同仁：都只中止它自己那一个监听器，
// 是否继续跑后面的由 breakOnError 决定。
func TestEmitBreakOnError(t *testing.T) {
	defer resetEvents()

	// breakOnError=true：panic 之后不再执行后续监听器
	resetEvents()
	var ran int
	On(EventTypStarted, func() error { panic("x") })
	On(EventTypStarted, func() error { ran++; return nil })
	if err := emit(EventTypStarted, true); err == nil {
		t.Fatal("应返回错误")
	}
	if ran != 0 {
		t.Error("breakOnError=true 时不该继续执行后续监听器")
	}

	// breakOnError=false：记日志但继续，最终返回 nil
	resetEvents()
	ran = 0
	On(EventTypStopped, func() error { panic("x") })
	On(EventTypStopped, func() error { ran++; return nil })
	if err := emit(EventTypStopped, false); err != nil {
		t.Errorf("best-effort 语义下应返回 nil, 实际 %v", err)
	}
	if ran != 1 {
		t.Error("breakOnError=false 时后续监听器仍应执行")
	}
}

// TestEventTypeString 事件名可读，不是裸数字
func TestEventTypeString(t *testing.T) {
	if EventTypLoaded.String() != "EventTypLoaded" {
		t.Errorf("应输出事件名, 实际 %v", EventTypLoaded.String())
	}
	if s := EventType(99).String(); !strings.Contains(s, "99") {
		t.Errorf("未知事件应保留数值, 实际 %v", s)
	}
}
