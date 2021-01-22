package cosnet

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/golang/protobuf/proto"
)

//TCP MESSAGE

var MsgHeadSize = 12
var MsgDataSize = 1024 * 1024

type MsgFlagType uint8
type MsgDataType uint8

const (
	MsgFlagEncrypt  MsgFlagType = 1 << 0 //数据是经过加密的
	MsgFlagCompress MsgFlagType = 1 << 1 //数据是经过压缩的
	MsgFlagContinue MsgFlagType = 1 << 2 //消息还有后续
	MsgFlagNeedAck  MsgFlagType = 1 << 3 //消息需要确认
	MsgFlagSubmit   MsgFlagType = 1 << 4 //确认消息
	MsgFlagReSend   MsgFlagType = 1 << 5 //重发消息
	MsgFlagClient   MsgFlagType = 1 << 6 //消息来自客服端，用于判断index来之服务器还是其他玩家
)

const (
	MsgDataTypeString MsgDataType = 1
	MsgDataTypeInt32  MsgDataType = 2
	MsgDataTypeJson   MsgDataType = 3
	MsgDataTypeProto  MsgDataType = 4
	MsgDataTypeXml    MsgDataType = 5
)

type Header struct {
	Size     int32  //数据长度 4294967295 4
	Index    int32  //消息序号 4
	Proto    uint16 //协议号 2
	Flags    MsgFlagType
	DataType MsgDataType //DATA 格式
}

type Message struct {
	Head *Header //消息头，可能为nil
	Data []byte  //消息数据
}

func NewMsg(b []byte) (*Message, error) {
	head := &Header{}
	if err := head.Parse(b); err != nil {
		return nil, err
	}
	return &Message{Head: head}, nil
}

//Bytes 生成成byte类型head
func (m *Header) Bytes() []byte {
	var b [][]byte
	b = append(b, IntToBytes(m.Size))
	b = append(b, IntToBytes(m.Index))
	b = append(b, IntToBytes(m.Proto))
	b = append(b, IntToBytes(m.DataType))
	b = append(b, IntToBytes(m.Flags))
	return bytes.Join(b, []byte{})
}

//parse 解析[]byte并填充字段
func (m *Header) Parse(head []byte) error {
	if len(head) != MsgHeadSize {
		return errors.New("head len error")
	}
	BytesToInt(head[0:4], &m.Size)
	BytesToInt(head[4:8], &m.Index)
	BytesToInt(head[8:10], &m.Proto)
	BytesToInt(head[10:11], &m.DataType)
	BytesToInt(head[11:12], &m.Flags)
	return nil
}

//Bytes 生成二进制文件
func (r *Message) Bytes() []byte {
	var b [][]byte
	b = append(b, r.Head.Bytes())
	if len(r.Data) > 0 {
		b = append(b, r.Data)
	}
	return bytes.Join(b, []byte{})
}

func (r *Message) NewMsg(proto uint16, data []byte) *Message {
	return &Message{
		Head: &Header{
			Size:  int32(len(data)),
			Index: r.Head.Index,
			Proto: proto,
		},
		Data: data,
	}
}

func (r *Message) Bind(i interface{}) error {
	dt := r.Head.DataType
	if dt == 0 {
		dt = Config.MsgDataType
	}
	switch dt {
	case MsgDataTypeJson:
		return json.Unmarshal(r.Data, i)
	case MsgDataTypeProto:
		return proto.Unmarshal(r.Data, i.(proto.Message))
	case MsgDataTypeXml:
		return xml.Unmarshal(r.Data, i.(proto.Message))
	default:
		return errors.New("unknown MsgDataType")
	}
}

func (m *MsgFlagType) Has(f MsgFlagType) bool {
	return (*m & f) > 0
}

func (m *MsgFlagType) Add(f MsgFlagType) {
	*m |= f
}
func (m *MsgFlagType) Del(f MsgFlagType) {
	if m.Has(f) {
		*m -= f
	}
}
