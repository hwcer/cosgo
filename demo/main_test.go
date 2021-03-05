package main

import (
	"testing"
)

var a = [10]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

func TestY(t *testing.T) {
	x := a[2:5]
	y := x[1:]

	y[1] = 10
	t.Log(a)
	t.Log(x)
	t.Log(y)
}
