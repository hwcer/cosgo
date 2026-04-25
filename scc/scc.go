package scc

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

// handle 处理函数类型，接收一个上下文参数
type handle func(context.Context)

// defaultCatchError 默认的错误捕获函数,将错误打印到控制台。
func defaultCatchError(err error) {
	fmt.Printf("%v", err)
}

// New 创建一个新的 SCC 实例。ctx 为 nil 时使用 context.Background()。
func New(ctx context.Context) *SCC {
	if ctx == nil {
		ctx = context.Background()
	}
	s := &SCC{Catch: defaultCatchError, WaitGroup: sync.WaitGroup{}}
	s.Context, s.cancel = context.WithCancel(ctx)
	s.WaitGroup.Add(1)
	return s
}

// SCC 协程控制器，用于管理和控制Go协程的生命周期
type SCC struct {
	sync.WaitGroup        // 嵌入的WaitGroup，用于等待所有协程结束
	stop    int32         // 停止标记，用于原子操作
	cancel  context.CancelFunc // 取消函数，用于取消所有子上下文
	Catch   func(error)   // 异常捕获函数，默认控制台打印
	Context context.Context // 根上下文
	handle  []func()      // 服务器关闭时执行的函数列表
}

// GO 启动一个普通的协程。
// 注意: Add(1) 必须在 go 之前执行,否则主线程 Wait 可能看到计数 0 提前返回。
func (s *SCC) GO(f func()) {
	s.WaitGroup.Add(1)
	go func() {
		defer s.WaitGroup.Done()
		f()
	}()
}

// CGO 启动一个带有取消通道的协程。
func (s *SCC) CGO(f handle) {
	s.WaitGroup.Add(1)
	go func() {
		defer s.WaitGroup.Done()
		ctx, cancel := s.WithCancel()
		defer cancel()
		f(ctx)
	}()
}

// SGO 启动一个使用 recover 保护的协程,防止主进程崩溃。
func (s *SCC) SGO(f handle) {
	s.WaitGroup.Add(1)
	go func() {
		defer s.WaitGroup.Done()
		defer func() {
			if e := recover(); e != nil {
				s.Catch(fmt.Errorf("%v\n%v", e, string(debug.Stack())))
			}
		}()
		ctx, cancel := s.WithCancel()
		defer cancel()
		f(ctx)
	}()
}

// Try 在当前 goroutine 同步执行 f,使用 recover 保护。
// 与 GO/CGO/SGO 不同,本函数不启动新 goroutine。
// WaitGroup 的 Add/Done 在同一线程内,无竞态。
func (s *SCC) Try(f handle) {
	defer func() {
		if e := recover(); e != nil {
			s.Catch(fmt.Errorf("%v\n%v", e, string(debug.Stack())))
		}
	}()
	s.WaitGroup.Add(1)
	defer s.WaitGroup.Done()
	ctx, cancel := s.WithCancel()
	defer cancel()
	f(ctx)
}

// Wait 阻塞等待所有协程结束,timeout=0 表示无限等待。
// 只能在主进程中调用;在 SCC 启动的协程内调用会自己等自己,死循环。
func (s *SCC) Wait(timeout time.Duration) (err error) {
	if timeout == 0 {
		s.WaitGroup.Wait()
	} else {
		err = s.Timeout(timeout, func() error {
			s.WaitGroup.Wait()
			return nil
		})
	}
	return
}

// Cancel 关闭所有协程。已关闭过返回 false,首次关闭返回 true。
func (s *SCC) Cancel() bool {
	if !atomic.CompareAndSwapInt32(&s.stop, 0, 1) {
		return false
	}
	s.WaitGroup.Done()
	s.cancel()
	for _, fn := range s.handle {
		fn()
	}
	return true
}

// Trigger 注册一个在服务器关闭时执行的函数。
// 调用假定:Trigger 在启动阶段单线程注册;Cancel 由 CAS 保证只执行一次。
// 已 Cancel 后再调用 Trigger 会被直接丢弃,不做追加(避免与 Cancel 遍历 handle 切片产生 race)。
func (s *SCC) Trigger(f func()) {
	if atomic.LoadInt32(&s.stop) != 0 {
		return
	}
	s.handle = append(s.handle, f)
}

// Stopped 返回是否已经关闭。
func (s *SCC) Stopped() bool {
	return s.stop > 0
}

// Deadline 返回上下文的截止时间。
func (s *SCC) Deadline() (deadline time.Time, ok bool) {
	return s.Context.Deadline()
}

// WithCancel 创建带取消功能的子上下文。
func (s *SCC) WithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(s.Context)
}

// WithTimeout 创建带超时功能的子上下文。
func (s *SCC) WithTimeout(t time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(s.Context, t)
}

// WithValue 基于 SCC 的根上下文创建带键值对的子上下文。
// 签名对齐 stdlib: 参数仅为 key/val。若需基于自定义父 context,
// 直接使用 context.WithValue(parent, key, val) 即可。
func (s *SCC) WithValue(key, val any) context.Context {
	return context.WithValue(s.Context, key, val)
}
