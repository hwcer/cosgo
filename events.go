package cosgo

import (
	"runtime/debug"

	"github.com/hwcer/logger"
)

var status EventType //当前状态
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

func emit(e EventType) {
	hs := events[e]
	if len(hs) == 0 {
		return
	}
	l := logger.LevelFATAL
	if e == EventTypReload {
		l = logger.LevelError
	}
	var err error
	defer func() {
		if s := recover(); s != nil {
			logger.Sprint(l, logger.Format(s), string(debug.Stack()))
		} else if err != nil {
			logger.Sprint(l, logger.Format(err))
		}
	}()
	
	for _, f := range hs {
		if err = f(); err != nil {
			return
		}
	}
}

func On(e EventType, f EventFunc) {
	events[e] = append(events[e], f)
}

func Status() EventType {
	return status
}
