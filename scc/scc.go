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

// defaultCatchError 默认的错误捕获函数，将错误打印到控制台
func defaultCatchError(err error) {
	fmt.Printf("%v", err)
}

// New 创建一个新的SCC实例
// @param ctx 上下文，如果为nil则使用context.Background()
// @return *SCC 返回创建的SCC实例
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

// GO 启动一个普通的协程
// @param f 要执行的函数
func (s *SCC) GO(f func()) {
	go func() {
		s.WaitGroup.Add(1)
		defer s.WaitGroup.Done()
		f()
	}()
}

// CGO 启动一个带有取消通道的协程
// @param f 要执行的处理函数，接收一个上下文参数
func (s *SCC) CGO(f handle) {
	go func() {
		s.WaitGroup.Add(1)
		defer s.WaitGroup.Done()
		ctx, cancel := s.WithCancel()
		defer cancel()
		f(ctx)
	}()
}

// SGO 启动一个使用recover保护的协程，防止主进程崩溃
// @param f 要执行的处理函数，接收一个上下文参数
func (s *SCC) SGO(f handle) {
	go func() {
		s.Try(f)
	}()
}

// Try 尝试执行一个处理函数，使用recover捕获异常
// @param f 要执行的处理函数，接收一个上下文参数
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

// Wait 阻塞模式等待所有协程结束
// 一般只在启动后主进程中使用
// 请不要在SCC创建的协程中使用，否者会无限等待
// @param timeout 超时时间，如果为0则无限等待
// @return error 错误信息，可能是超时错误
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

// Cancel 关闭所有协程
// @return bool 是否成功关闭，如果已经关闭过则返回false
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

// Trigger 注册一个在服务器关闭时执行的函数
// @param f 要执行的函数
func (s *SCC) Trigger(f func()) {
	s.handle = append(s.handle, f)
}

// Stopped 判断是否已经关闭
// @return bool 是否已关闭
func (s *SCC) Stopped() bool {
	return s.stop > 0
}

// Deadline 返回上下文的截止时间
// @return deadline 截止时间
// @return ok 是否有截止时间
func (s *SCC) Deadline() (deadline time.Time, ok bool) {
	return s.Context.Deadline()
}

// WithCancel 创建一个带有取消功能的子上下文
// @return context.Context 创建的子上下文
// @return context.CancelFunc 取消函数
func (s *SCC) WithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(s.Context)
}

// WithTimeout 创建一个带有超时功能的子上下文
// @param t 超时时间
// @return context.Context 创建的子上下文
// @return context.CancelFunc 取消函数
func (s *SCC) WithTimeout(t time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(s.Context, t)
}

// WithValue 创建一个带有键值对的子上下文
// @param parent 父上下文，如果为nil则使用SCC的根上下文
// @param key 键
// @param val 值
// @return context.Context 创建的子上下文
func (s *SCC) WithValue(parent context.Context, key, val any) context.Context {
	if parent == nil {
		return context.WithValue(s.Context, key, val)
	} else {
		return context.WithValue(parent, key, val)
	}
}
