package utils

import (
	"fmt"
	"testing"
)

func TestBitwis_Add(t *testing.T) {
	var x Bitwise
	fmt.Printf("Bitwise:%b\n", x)

	x.Add(2)
	fmt.Printf("Bitwise Add:%b\n", x)

	fmt.Printf("Bitwise Has:%v\n", x.Has(2))

	x.Del(2)
	fmt.Printf("Bitwise Del:%b\n", x)
	fmt.Printf("Bitwise Has:%v\n", x.Has(2))

	x.Add(65)
	fmt.Printf("Bitwise Add:%b\n", x)

	x.Del(65)
	fmt.Printf("Bitwise Del:%b\n", x)
}
