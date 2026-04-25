package cosgo

import (
	"fmt"

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

var events map[EventType][]EventFunc

func init() {
	events = make(map[EventType][]EventFunc)
}

// emit 触发事件 e 的所有监听器。
//   - breakOnError=true: 首个返回 error 的监听器使整条链短路,返回该 error。
//   - breakOnError=false: 每个 error 被记入日志但不中断,最终返回 nil(best-effort 语义)。
//   - 任意监听器 panic 都会被 recover,转换为 error 返回(优先级高于 f() 的 error)。
func emit(e EventType, breakOnError bool) (err error) {
	hs := events[e]
	if len(hs) == 0 {
		return
	}
	l := logger.LevelFatal
	if e == EventTypReload {
		l = logger.LevelError
	}
	defer func() {
		if s := recover(); s != nil {
			err = fmt.Errorf("cosgo emit[%d] error: %v", e, s)
		}
	}()

	for _, f := range hs {
		if err = f(); err != nil {
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
