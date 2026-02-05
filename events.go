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
