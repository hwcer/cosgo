package xnet

import (
	"errors"
	"unsafe"
)


const MsgHeadSize = 12

var MaxMsgDataSize uint32 = 1024 * 1024

type MsgFlagType uint16
const (
	MsgFlagEncrypt   MsgFlagType   = 1 << 0 //数据是经过加密的
	MsgFlagCompress  MsgFlagType  = 1 << 1 //数据是经过压缩的
	MsgFlagContinue  MsgFlagType   = 1 << 2 //消息还有后续
	MsgFlagNeedAck   MsgFlagType  = 1 << 3 //消息需要确认
	MsgFlagSubmit    MsgFlagType  = 1 << 4 //确认消息
	MsgFlagReSend    MsgFlagType  = 1 << 5 //重发消息
	MsgFlagClient    MsgFlagType  = 1 << 6 //消息来自客服端，用于判断index来之服务器还是其他玩家
)



type MsgHead struct {
	Len   	uint32 //数据长度 4294967295 4
	Act   	uint16 //协议号 65535   2
	Index   uint32  //消息报序号 2
	Flags   uint16 //标记    2
}

type Message struct {
	Head      *MsgHead //消息头，可能为nil
	Data       []byte       //消息数据
}





//Bytes 生成成byte类型head
func (m *MsgHead) Bytes() []byte {
	data := make([]byte, MsgHeadSize)
	phead := (*MsgHead)(unsafe.Pointer(&data[0]))
	phead.Len = m.Len
	phead.Act = m.Act
	phead.Index = m.Index
	phead.Flags = m.Flags
	return data
}
//parse 解析[]byte并填充字段
func (m *MsgHead) FromBytes(head []byte) error{
	if len(head) != MsgHeadSize  {
		return errors.New("head len error")
	}
	phead := (*MsgHead)(unsafe.Pointer(&head[0]))
	m.Len = phead.Len
	m.Act = phead.Act
	m.Index = phead.Index
	m.Flags = phead.Flags
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

func (m *MsgHead) HasFlag(f MsgFlagType)bool{
	return  (m.Flags&uint16(f)) > 0
}

func (m *MsgHead) AddFlag(f MsgFlagType){
	m.Flags |= uint16(f)
}
func (m *MsgHead) SubFlag(f MsgFlagType){
	m.Flags -= uint16(f)
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

func NewMsg(act uint16, data []byte) *Message {
	return &Message{
		Head: &MsgHead{
			Len:   uint32(len(data)),
			Act:   act,
		},
		Data: data,
	}
}
func NewMsgData(data []byte) *Message {
	return &Message{
		Head: &MsgHead{
			Len: uint32(len(data)),
		},
		Data: data,
	}
}
