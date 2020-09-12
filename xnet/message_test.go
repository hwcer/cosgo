package xnet

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestMessage(t *testing.T) {
	fmt.Println("MsgHeadSize", MsgHeadSize)

	m :=NewMsg(110,[]byte("test"))
	m.Head.Act = 65535
	//m.Head.AddFlag(MsgFlagEncrypt)
	//m.Head.AddFlag(MsgFlagCompress)
	//m.Head.AddFlag(MsgFlagContinue)
	//m.Head.AddFlag(MsgFlagNeedAck)
	//m.Head.AddFlag(MsgFlagSubmit)
	//m.Head.AddFlag(MsgFlagReSend)
	//m.Head.AddFlag(MsgFlagClient)
	m.Head.Flags |= uint16(1 << 15)

	fmt.Printf("msg Head:%+v  Data:%v\n", m.Head,m.Data)
	b := m.Bytes()
	fmt.Printf("msg b:%+v\n", b)

	h := m.Head.Bytes()
	fmt.Printf("msg h:%+v\n", h)

	fmt.Printf("msg Head Len size:%v\n", unsafe.Sizeof(m.Head.Len))
	fmt.Printf("msg Head Act size:%v\n", unsafe.Sizeof(m.Head.Act))
	fmt.Printf("msg Head Index size:%v\n", unsafe.Sizeof(m.Head.Index))
	fmt.Printf("msg Head Flags size:%v\n", unsafe.Sizeof(m.Head.Flags))


	newMsgHead := NewMsgHead(h)
	fmt.Printf("newMsgHead:%+v\n", newMsgHead)
	fmt.Printf("newMsgHead:%+v\n", newMsgHead.Bytes())

}
