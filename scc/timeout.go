package scc

import (
	"errors"
	"time"
)

var ErrorTimeout = errors.New("timeout")

func Timeout(d time.Duration, fn func() error) error {
	return scc.Timeout(d, fn)
}

func (s *SCC) Timeout(d time.Duration, fn func() error) error {
	cher := make(chan error)
	go func() {
		cher <- fn()
	}()
	select {
	case err := <-cher:
		return err
	case <-time.After(d):
		return ErrorTimeout
	}
}
