package utils

import (
	"fmt"
	"testing"
)

func TestBitwis_Add(t *testing.T) {
	var x Bitwise
	fmt.Printf("Bitwise:%b\n", x)

	x.Set(2)
	fmt.Printf("Bitwise connect:%b\n", x)

	fmt.Printf("Bitwise Has:%v\n", x.Has(2))

	x.Delete(2)
	fmt.Printf("Bitwise destroy:%b\n", x)
	fmt.Printf("Bitwise Has:%v\n", x.Has(2))

	x.Set(3)
	fmt.Printf("Bitwise connect:%b\n", x)

	x.Delete(3)
	fmt.Printf("Bitwise destroy:%b\n", x)

	x.Delete(4)
	fmt.Printf("Bitwise destroy:%b\n", x)
}
