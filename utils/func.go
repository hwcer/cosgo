package utils

import (
	"errors"
	"fmt"
	"math"
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

// Ceil 除法向上取整
func Ceil(a, b int32) int32 {
	r := a / b
	if a%b != 0 {
		r += 1
	}
	return r
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
	for i := 0; i < len(s); i++ {
		if v := s[i]; v < 32 || v == 127 { //空格32
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

// FloatPrecision 四舍五入，保留到Precision位小数
func FloatPrecision(value float64, precision float64) float64 {
	x := math.Pow(10, precision)
	return math.Round(value*x) / x
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
