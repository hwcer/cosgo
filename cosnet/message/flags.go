package message

type Flags uint8
type ContentType uint8

const (
	ContentTypeNumber ContentType = 1
	ContentTypeString ContentType = 2
	ContentTypeJson   ContentType = 3
	ContentTypeXml    ContentType = 4
	ContentTypeProto  ContentType = 5
)


const (
	MsgFlagEncrypt  Flags = 1 << 0 //数据是经过加密的
	MsgFlagCompress Flags = 1 << 1 //数据是经过压缩的
	MsgFlagContinue Flags = 1 << 2 //消息还有后续
	MsgFlagNeedAck  Flags = 1 << 3 //消息需要确认
	MsgFlagSubmit   Flags = 1 << 4 //确认消息
	MsgFlagReSend   Flags = 1 << 5 //重发消息
	MsgFlagClient   Flags = 1 << 6 //消息来自客服端，用于判断index来之服务器还是其他玩家
)

func (m *Flags) Has(f Flags) bool {
	return (*m & f) > 0
}

func (m *Flags) Add(f Flags) {
	*m |= f
}
func (m *Flags) Del(f Flags) {
	if m.Has(f) {
		*m -= f
	}
}
