package utils

import (
	"errors"
	"fmt"
	"runtime/debug"
	"time"
	"unicode"
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

func IncludeNotPrintableChar(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) || (r >= 0 && r <= 31) || (r == 160) || (r >= 8232 && r <= 8233) {
			return true
		}
	}
	return false
}

func Assert(ps ...func() error) (err error) {
	for _, p := range ps {
		if err = p(); err != nil {
			return err
		}
	}
	return nil
}

func UTF8StringLen(str string) int {
	l := 0
	for _, v := range str {
		if v > 0xFF {
			l += 2
		} else {
			l += 1
		}
	}
	return l
}

func CloneMap[T1 comparable, T2 comparable](src map[T1]T2) map[T1]T2 {
	r := map[T1]T2{}
	for k, v := range src {
		r[k] = v
	}
	return r
}

func MapKeys[T1 comparable, T2 comparable](src map[T1]T2) []T1 {
	r := make([]T1, 0, len(src))
	for k, _ := range src {
		r = append(r, k)
	}
	return r
}
