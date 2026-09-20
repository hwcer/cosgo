// Package phase 应用启动阶段的唯一事实源。
//
// 记录"当前启动到哪一步"(Init→Starting→Started→Closing→Closed),由 cosgo.Start/stop
// 驱动推进。各模块"仅启动期注册"类 API(事件表、保留键黑名单等)的封板守卫直接读
// 这里——新增需要封板的模块时,调 phase.Sealed() 判断即可,不再依赖根包在 Start 里
// 级联调用各自的 Seal* 函数。
//
// 🔴 本包是叶子依赖:只允许 import logger,被 cosgo 根包/session/任意下游模块引用,
// 不许反向依赖任何业务包(否则会构成 import 环)。
package phase

import (
	"sync/atomic"

	"github.com/hwcer/logger"
)

// Type 应用启动阶段
type Type int32

const (
	Init     Type = iota //构造期/未启动
	Starting             //Start 执行中:模块逐个 Init/Start,监听类模块已可受理请求
	Started              //启动完成——🔴 封板点:仅启动期注册自此拒绝
	Closing              //关闭中
	Closed               //已关闭
)

var names = map[Type]string{
	Init:     "Init",
	Starting: "Starting",
	Started:  "Started",
	Closing:  "Closing",
	Closed:   "Closed",
}

func (t Type) String() string {
	if s, ok := names[t]; ok {
		return s
	}
	return "Phase(" + itoa(int32(t)) + ")"
}

func itoa(v int32) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var b [12]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

var current atomic.Int32

// Set 推进阶段(由 cosgo.Start/stop 驱动;业务模块一般只读不写)
func Set(t Type) {
	current.Store(int32(t))
}

// Get 当前阶段
func Get() Type {
	return Type(current.Load())
}

// Is 当前是否恰好处于阶段 t
func Is(t Type) bool {
	return Get() == t
}

// Sealed 是否已越过封板点(Started/Closing/Closed 均视为已封板,关闭期的注册
// 同样必须拒绝):各模块"仅启动期注册"的统一守卫。
//
// ⚠️ 封板点在 cosgo.Start 的模块全部 Start、EventTypStarted 发完之后——模块 Start
// 已开始受理请求而封板尚未发生的毫秒级尾窗内,Emit×On 并发写裸 map 仍是 fatal
// 竞争,契约依旧是"监听器只允许在包 init 或 Module.Init 期注册",本包只是把
// 拦截时刻做成全局一致的时钟,并不放宽注册窗口
func Sealed() bool {
	return Get() >= Started
}

// Alert 封板后的统一拒绝提示:各模块守卫复用同一文案格式,日志里可一眼识别
// "运行期操作被忽略"这一类误用
func Alert(api string, args ...any) {
	logger.Alert("%v after server started(phase=%v): 仅启动期操作,本次调用已忽略", append([]any{api, Get()}, args...)...)
}
