package message

type Flags uint8

const (
	FlagEncrypt  Flags = 1 << 0 //数据是经过加密的
	FlagCompress Flags = 1 << 1 //数据是经过压缩的
	FlagContinue Flags = 1 << 2 //消息还有后续
	FlagNeedAck  Flags = 1 << 3 //消息需要确认
	FlagSubmit   Flags = 1 << 4 //确认消息
	FlagReSend   Flags = 1 << 5 //重发消息
	FlagClient   Flags = 1 << 6 //消息来自客服端，用于判断index来之服务器还是其他玩家
)

func (m *Flags) Has(f Flags) bool {
	return (*m & f) > 0
}

func (m *Flags) Set(f Flags) {
	*m |= f
}
func (m *Flags) Remove(f Flags) {
	if m.Has(f) {
		*m -= f
	}
}
