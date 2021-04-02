package cosnet

import (
	"fmt"
	"testing"
)

func TestSocket(t *testing.T) {
	var a ObjectID
	x := uint32(1024)
	y := uint32(88888)
	a.Pack(x, y)

	fmt.Printf("a:%v\n", a)
	b := a.Parse()
	fmt.Printf("a:%v,b:%v\n", a, b)

	fmt.Printf("x:%b\n", x)
	fmt.Printf("y:%b\n", y)
	fmt.Printf("a:%b\n", a)
}
