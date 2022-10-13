package utils

import (
	"fmt"
	"time"
)

func Try(f func(), handler ...func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			if len(handler) == 0 {
				fmt.Printf("%v", err)
			} else {
				handler[0](err)
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
		return ErrorTimeout
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
