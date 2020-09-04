package network

import (
	"fmt"
	"testing"
	"unsafe"
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

	fmt.Printf("msg Head Len size:%v\n", unsafe.Sizeof(m.Head.Len))
	fmt.Printf("msg Head Act size:%v\n", unsafe.Sizeof(m.Head.Act))
	fmt.Printf("msg Head Sid size:%v\n", unsafe.Sizeof(m.Head.Sid))
	fmt.Printf("msg Head Flags size:%v\n", unsafe.Sizeof(m.Head.Flags))


	newMsgHead := NewMsgHead(h)
	fmt.Printf("newMsgHead:%+v\n", newMsgHead)
	fmt.Printf("newMsgHead:%+v\n", newMsgHead.Bytes())


	//土法转换
	headByte := newMsgHead.Bytes()
	head := MsgHead{}
	head.Len = BytesToUint32(headByte[0:4])
	head.Act = BytesToUint16(headByte[4:6])
	head.Sid = BytesToUint16(headByte[6:8])
}

func BytesToUint16(b []byte) uint16 {
	blen := len(b)
	if blen > 4 {
		blen = 4
	}
	v := uint16(0)
	for i := 0; i < blen; i++ {
		v += uint16(b[i]) << uint16(8*i)
	}
	return v
}

func BytesToUint32(b []byte) uint32 {
	blen := len(b)
	if blen > 4 {
		blen = 4
	}
	v := uint32(0)
	for i := 0; i < blen; i++ {
		v += uint32(b[i]) << uint32(8*i)
	}
	return v
}

func Uint32ToBytes(value uint32) []byte {
	b := make([]byte, 4)
	for i := 0; i < 4; i++ {
		b[i] = byte(value >> uint32(8*i))
	}
	return b
}

