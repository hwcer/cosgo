package main

import (
	"testing"
)

var a []int

func init() {
	for i := 0; i <= 10; i++ {
		a = append(a, i)
	}
}

func BenchmarkAppend(t *testing.B) {
	var b []int
	for i := 0; i <= 100000; i++ {
		b = append([]int{}, a...)
	}
	b = b
}

func BenchmarkCopy(t *testing.B) {
	b := make([]int, len(a))
	for i := 0; i <= 100000; i++ {
		copy(b, a)
	}
}

func BenchmarkX(t *testing.B) {
	var b []int
	for i := 0; i <= 100000; i++ {
		b = a[0:]
	}
	b = b
}

func TestY(t *testing.T) {
	b := a
	b = b[1:]
	b = b[1:]
	b = b[1:]
	b = b[1:]
	b = b[1:]
	t.Logf("a %v", a)
	t.Logf("b %v", b)
}
