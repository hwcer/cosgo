package cosgo

type EventType int32
type EventFunc func() error

const (
	EventTypStart   EventType = iota //开始启动
	EventTypLoader                   //(Init)加载完成
	EventTypReady                    //启动完成
	EventTypClosing                  //开始关闭
	EventTypStopped                  //停止之后
)

var events map[EventType][]EventFunc

func init() {
	events = make(map[EventType][]EventFunc)
}

func emit(e EventType) (err error) {
	if hs, ok := events[e]; ok {
		for _, f := range hs {
			if err = f(); err != nil {
				return
			}
		}
	}
	return
}

func On(e EventType, f EventFunc) {
	events[e] = append(events[e], f)
}
