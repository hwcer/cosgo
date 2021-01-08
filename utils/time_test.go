package utils

import (
	"net"
	"testing"
)

func TestGetDayStartTime(t *testing.T) {
	addr, err := net.ResolveTCPAddr("tcp", "192.168.1.1:80")
	if err != nil {
		t.Log("错误：", err)
	} else {
		t.Log(addr.Network(), addr.IP, addr.Port, "-----", addr.String())
	}
	ip := addr.IP.To16()
	t.Log(ip.String())
	//addr, err := net.ResolveIPAddr("ip", "192.168.1.1")
	//if err != nil {
	//	t.Log("错误：", err)
	//} else {
	//	t.Log(addr.Network(), addr.IP)
	//}

	//ip := net.ParseIP("192.168.0.1")
	//ip = ip.To16()
	//t.Log(ip.String())

}
