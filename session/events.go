package session

import (
	"sync/atomic"

	"github.com/hwcer/logger"
)

type Event int8
type Listener func(any)

const (
	EventSessionNew        Event = iota //SESSION New,参数 *Data
	EventSessionCreated                 //SESSION Create时,参数 *Data
	EventSessionRelease                 //销毁SESSION时,参数 *Data（超时后触发）
	EventSessionDisconnect              //SESSION 掉线时,参数 *Data（连接断开立即触发，早于Release）
	EventHeartbeat                      //心跳,参数 心跳间隔 int32
)

// listeners 事件订阅表:仅启动期注册(业务在包 init 或启动钩子里注册),
// Cosgo.Start 完成后由根包调用 SealEvents 封板,封板后再 On 只提示不注册(见 On)。
// 🔴 契约化的理由同根 events:运行期注册会与 Emit 构成 concurrent map fatal
// (不可恢复)故封板拦下;注册期单线程,裸 map 零成本
var (
	listeners       map[Event][]Listener
	listenersSealed atomic.Bool
)

func init() {
	listeners = make(map[Event][]Listener)
}

// SealEvents 封板事件表(由 cosgo.Start 在启动完成后调用,业务无需手动调用)
func SealEvents() {
	listenersSealed.Store(true)
}

// On 注册事件监听器(仅启动期)。封板后调用只提示不注册——不崩进程:
// 运行期注册会与 Emit 构成 concurrent map fatal,拦下即可;误用多来自
// 重连/热更路径的旧代码,Alert 进日志留排查线索。
func On(event Event, listener Listener) {
	if listenersSealed.Load() {
		logger.Alert("session.On(%v) after server started: 事件监听器仅允许在启动期注册,本次注册已忽略", event)
		return
	}
	listeners[event] = append(listeners[event], listener)
}

// Emit 触发事件(注册期封板后运行期只读,无锁遍历)
func Emit(event Event, value any) {
	for _, l := range listeners[event] {
		l(value)
	}
}

// Listen 是 On 的别名。
func Listen(event Event, listener Listener) {
	On(event, listener)
}
