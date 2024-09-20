package await

import (
	"sync/atomic"
)

func NewInitialize(handle func() error) *Initialize {
	i := new(Initialize)
	i.handle = handle
	i.initializing = make(chan struct{})
	return i
}

// Initialize 并发模式下初始化控制器
type Initialize struct {
	status       int32
	handle       func() error //如果没有初始化则进行初始化
	initializing chan struct{}
}

func (i *Initialize) Verify() error {
	if i.status >= 2 {
		return nil
	}
	if !atomic.CompareAndSwapInt32(&i.status, 0, 1) {
		<-i.initializing
		return nil
	}
	defer func() {
		i.status = 2
		close(i.initializing)
	}()
	return i.handle()
}

func (i *Initialize) Reload() {
	if atomic.CompareAndSwapInt32(&i.status, 2, 3) {
		i.initializing = make(chan struct{})
		i.status = 0
	}
}
