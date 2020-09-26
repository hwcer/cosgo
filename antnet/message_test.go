package antnet

import (
	"fmt"
	"testing"
)

func TestMessage(t *testing.T) {
	fmt.Println("MsgHeadSize", MsgHeadSize)

	m := NewMsg(110, []byte("test"))
	m.Head.Index = 456
	m.Head.Proto = 12399
	for i := 0; i <= 16; i++ {
		m.Head.AddFlag(MsgFlagType(1 << i))
	}

	fmt.Printf("msg Head:%+v  Data:%v\n", m.Head, m.Data)
	b := m.Bytes()
	fmt.Printf("msg b:%+v\n", b)

	h := m.Head.Bytes()
	fmt.Printf("msg head len:%+v\n", len(h))
	fmt.Printf("msg h:%+v\n", h)

	newMsgHead := NewMsgHead(h)
	fmt.Printf("newMsgHead:%+v\n", newMsgHead)
	fmt.Printf("newMsgHead:%+v\n", newMsgHead.Bytes())

}
