package network

import (
	"errors"
	"unsafe"
)
const (
	FlagEncrypt  = 1 << 0 //数据是经过加密的
	FlagCompress = 1 << 1 //数据是经过压缩的
	FlagContinue = 1 << 2 //消息还有后续
	FlagNeedAck  = 1 << 3 //消息需要确认
	FlagAck      = 1 << 4 //确认消息
	FlagReSend   = 1 << 5 //重发消息
	FlagClient   = 1 << 6 //消息来自客服端，用于判断index来之服务器还是其他玩家
)

type MsgHead struct {
	Len   	uint32 //数据长度
	Act   	uint16 //客户端动作
	Sid   	uint16 //服务器发包序号
	Flags   uint16 //标记
}

type Message struct {
	Head      *MsgHead //消息头，可能为nil
	Data       []byte       //消息数据
}


type MsgSliceMock struct {
	addr uintptr
	len  int
	cap  int
}

const MsgHeadSize = int(unsafe.Sizeof(MsgHead{}))

//Bytes 生成成byte类型head
func (m *MsgHead) Bytes() []byte {
	msgBytes := &MsgSliceMock{
		addr: uintptr(unsafe.Pointer(m)),
		cap:  MsgHeadSize,
		len:  MsgHeadSize,
	}
	data := *(*[]byte)(unsafe.Pointer(msgBytes))
	return data
}
//parse 解析[]byte并填充字段
func (m *MsgHead) FromBytes(head []byte) error{
	if len(head) != MsgHeadSize  {
		return errors.New("head len error")
	}
	r := *(**MsgHead)(unsafe.Pointer(&head))
	*m = *r
	return nil
}
//按照Message.Data长度设置head.Len并返回[]byte类型的Message
func (m *MsgHead) BytesWithData(data []byte) []byte {
	m.Len = uint32(len(data))
	msgLen := MsgHeadSize + int(m.Len)
	msg := make([]byte,msgLen,msgLen)
	copy(msg,m.Bytes())
	if m.Len>0 {
		copy(msg[MsgHeadSize:], data)
	}
	return msg
}



//Bytes 生成二进制文件
func (r *Message) Bytes() []byte {
	if r.Head != nil {
		if len(r.Data) >0 {
			return r.Head.BytesWithData(r.Data)
		}else{
			return r.Head.Bytes()
		}
	}
	return r.Data
}


func NewMsgHead(head []byte) *MsgHead {
	msg := &MsgHead{}
	if err := msg.FromBytes(head);err!=nil{
		return nil
	}
	return msg
}

func NewMsgData(data []byte) *Message {
	return &Message{
		Head: &MsgHead{
			Len: uint32(len(data)),
		},
		Data: data,
	}
}

func NewMsg(act uint16, data []byte) *Message {
	return &Message{
		Head: &MsgHead{
			Len:   uint32(len(data)),
			Act:   act,
		},
		Data: data,
	}
}
