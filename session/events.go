package session

import (
	"sync"
	"sync/atomic"
)

type Event int8
type Listener func(any)

const (
	EventSessionNew     Event = iota //SESSION New,参数 *Data
	EventSessionCreated              //SESSION Create时,参数 *Data
	EventSessionRelease              //销毁SESSION时,参数 *Data
	EventHeartbeat                   //心跳,参数 心跳间隔 int32
)

// listeners 事件订阅表,Copy-on-Write 发布:
//   - Emit 路径无锁,atomic.Load 当前快照后遍历,读到的永远是一致的完整快照。
//   - On 路径用 listenersMu 串行写者,拷贝整张 map + 拷贝对应事件的 slice,
//     再 atomic.Store 发布。旧快照被其它读者持有时继续有效。
//
// 适用场景:订阅少、触发多(典型的事件模型)。
var (
	listenersMu sync.Mutex
	listenersV  atomic.Pointer[map[Event][]Listener]
)

func init() {
	empty := map[Event][]Listener{}
	listenersV.Store(&empty)
}

// On 注册事件监听器。写路径:拷贝旧 map,为目标事件创建全新 slice 并追加,再原子发布。
func On(event Event, listener Listener) {
	listenersMu.Lock()
	defer listenersMu.Unlock()
	old := *listenersV.Load()
	next := make(map[Event][]Listener, len(old)+1)
	for k, v := range old {
		next[k] = v
	}
	// 强制为目标事件新建 backing array,避免对旧 slice 的共享 append 破坏其它读者
	prev := old[event]
	nslice := make([]Listener, len(prev)+1)
	copy(nslice, prev)
	nslice[len(prev)] = listener
	next[event] = nslice
	listenersV.Store(&next)
}

// Emit 触发事件,无锁读取当前快照。
func Emit(event Event, value any) {
	m := *listenersV.Load()
	for _, l := range m[event] {
		l(value)
	}
}

// Listen 是 On 的别名。
func Listen(event Event, listener Listener) {
	On(event, listener)
}
