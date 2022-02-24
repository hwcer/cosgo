package cosnet

var Options = struct {
	WriteChanSize       int32  //写通道缓存
	ConnectMaxSize      int32  //连接人数
	ClientReconnectMax  uint16 //断线重连最大尝试次数
	ClientReconnectTime uint16 //断线重连每次等待时间(s) ClientReconnectTime * ReconnectNum
	SocketHeartbeat     uint16 //(MS)服务器心跳,用来检测玩家僵尸连接
	SocketConnectTime   uint16 //[基于心跳间隔]连接超时几次心跳没有动作被判断掉线
	AutoCompressSize    uint32 //自动压缩
}{
	WriteChanSize:       500,
	ConnectMaxSize:      50000,
	ClientReconnectMax:  1000,
	ClientReconnectTime: 5,
	SocketHeartbeat:     1000,
	SocketConnectTime:   10,
}
