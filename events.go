package cosgo

import (
	"fmt"
	"reflect"
	"runtime"
	"runtime/debug"

	"github.com/hwcer/logger"
)

type EventType int32
type EventFunc func() error

const (
	EventTypBegin   EventType = iota //开始启动
	EventTypLoaded                   //(Init)加载完成
	EventTypStarted                  //启动完成
	EventTypClosing                  //开始关闭
	EventTypStopped                  //停止之后
	EventTypReload                   //reload
)

var eventNames = map[EventType]string{
	EventTypBegin:   "EventTypBegin",
	EventTypLoaded:  "EventTypLoaded",
	EventTypStarted: "EventTypStarted",
	EventTypClosing: "EventTypClosing",
	EventTypStopped: "EventTypStopped",
	EventTypReload:  "EventTypReload",
}

func (e EventType) String() string {
	if s, ok := eventNames[e]; ok {
		return s
	}
	return fmt.Sprintf("EventType(%d)", int32(e))
}

var events map[EventType][]EventFunc

func init() {
	events = make(map[EventType][]EventFunc)
}

// funcName 取监听器的函数名,形如 server/game/handle/trial.seed
//
// 这是排查启动失败最关键的一条信息:事件监听器是各模块在 init 里 On() 进来的,
// 出错时调用栈只剩 cosgo.Start → main.main,完全看不出是谁注册的。
func funcName(f EventFunc) string {
	if f == nil {
		return "<nil>"
	}
	p := reflect.ValueOf(f).Pointer()
	if fn := runtime.FuncForPC(p); fn != nil {
		name := fn.Name()
		if file, line := fn.FileLine(p); file != "" {
			return fmt.Sprintf("%s (%s:%d)", name, file, line)
		}
		return name
	}
	return fmt.Sprintf("func@%#x", p)
}

// invoke 执行单个监听器,把 panic 转成带【原始堆栈】的 error。
//
// 🔴 recover 必须【逐个监听器】做,不能在 emit 外面包一层:
//   - 包在外面只知道"这个事件炸了",不知道是哪个监听器——而监听器是各模块
//     在自己的 init 里 On() 进来的,顺序对使用者不可见。
//   - debug.Stack() 必须在 recover 的那个 defer 里取。等 emit 返回后再取,
//     栈早已回卷,只剩 cosgo.Start → main.main,业务层帧全没了
//     (2026-08-21 的启动崩溃就是这个症状:一个 nil 解引用,日志里完全看不出在哪)。
func invoke(e EventType, i int, f EventFunc) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v 第 %d 个监听器 panic\n  监听器: %s\n  原因: %v\n%s",
				e, i, funcName(f), r, debug.Stack())
		}
	}()
	if err = f(); err != nil {
		//返回 error 的路径同样要指名道姓,否则一样查不出是谁
		err = fmt.Errorf("%v 第 %d 个监听器返回错误\n  监听器: %s\n  原因: %w",
			e, i, funcName(f), err)
	}
	return
}

// emit 触发事件 e 的所有监听器。
//   - breakOnError=true: 首个返回 error 的监听器使整条链短路,返回该 error。
//   - breakOnError=false: 每个 error 被记入日志但不中断,最终返回 nil(best-effort 语义)。
//   - 任意监听器 panic 都会被 recover 成 error,**带上原始堆栈与监听器函数名**。
//     panic 只中止它自己那一个监听器;是否继续跑后面的由 breakOnError 决定,
//     与"返回 error"一视同仁。
func emit(e EventType, breakOnError bool) (err error) {
	hs := events[e]
	if len(hs) == 0 {
		return
	}
	l := logger.LevelFatal
	if e == EventTypReload {
		l = logger.LevelError
	}
	for i, f := range hs {
		if err = invoke(e, i, f); err != nil {
			if breakOnError {
				return err
			}
			logger.Sprint(l, logger.Format(err))
		}
	}
	return nil
}

func On(e EventType, f EventFunc) {
	events[e] = append(events[e], f)
}
