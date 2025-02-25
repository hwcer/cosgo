package utils

import (
	"fmt"
	"testing"
)

func TestAddress(t *testing.T) {
	ip := "192.168.1.2:8000"

	code := IPv4Encode(ip)
	fmt.Printf("ip Encode:%v\n", code)
	addr := IPv4Decode(code)
	fmt.Printf("ip Decode:%v\n", addr)
}
