package utils

import (
	"errors"
	"fmt"
	"runtime/debug"
	"time"
)

type TryHandle func(any)

func Try(f func(), handle ...TryHandle) {
	defer func() {
		if err := recover(); err != nil {
			if len(handle) == 0 {
				fmt.Printf("%v\n%v", err, string(debug.Stack()))
			} else {
				handle[0](err)
			}
		}
	}()
	f()
}

func Timeout(d time.Duration, fn func() error) error {
	cher := make(chan error)
	go func() {
		cher <- fn()
	}()
	select {
	case err := <-cher:
		return err
	case <-time.After(d):
		return errors.New("timeout")
	}
}

func Sprintf(format interface{}, args ...interface{}) (r string) {
	switch v := format.(type) {
	case string:
		if len(args) > 0 {
			r = fmt.Sprintf(v, args...)
		} else {
			r = v
		}
	default:
		r = fmt.Sprintf("%v", format)
	}
	return
}
