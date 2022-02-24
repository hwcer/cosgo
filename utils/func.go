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


