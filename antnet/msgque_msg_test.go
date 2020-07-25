package antnet

import (
	"fmt"
	"testing"
)

func TestMessage(t *testing.T) {
	fmt.Println("MsgHeadSize", MsgHeadSize)

	m :=NewMsg(110,[]byte("test"))
	m.Head.Sid = 10
	fmt.Printf("msg Head:%+v  Data:%v\n", m.Head,m.Data)
	b := m.Bytes()
	fmt.Printf("msg b:%+v\n", b)

	h := m.Head.Bytes()
	fmt.Printf("msg h:%+v\n", h)

	newMsg := NewMsgHead(h)
	fmt.Printf("newMsg:%+v\n", newMsg)
	fmt.Printf("newMsg:%+v\n", newMsg.Bytes())
}
