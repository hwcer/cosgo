package scc

import (
	"errors"
	"time"
)

var ErrorTimeout = errors.New("timeout")

// Timeout 在 d 时间内等待 fn 完成,超时返回 ErrorTimeout。
// cher 为带 1 缓冲的 channel:即使超时分支获胜,子 goroutine 的 send 也能立即完成并退出,
// 不会因无接收者而永久阻塞造成 goroutine 泄漏。
func Timeout(d time.Duration, fn func() error) error {
	cher := make(chan error, 1)
	go func() {
		cher <- fn()
	}()
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case err := <-cher:
		return err
	case <-timer.C:
		return ErrorTimeout
	}
}

func (s *SCC) Timeout(d time.Duration, fn func() error) error {
	return Timeout(d, fn)
}
