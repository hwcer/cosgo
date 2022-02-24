package cosnet

type NetType uint8
type EventType uint8

//NetType
const (
	NetworkTcp    NetType = 1      //TCP
	NetworkUdp            = 1 << 1 //UDP
	NetworkWss            = 1 << 2 //wss||ws
	NetTypeClient         = 1 << 3 //client Request
	NetTypeServer         = 1 << 4 //Server Listener
)

//EventType
const (
	EventTypeHeartbeat  EventType = iota + 1 //心跳事件
	EventTypeConnected                       //连接成功
	EventTypeDisconnect                      //断开连接
)

//常用NetType组合
const (
	AnyNetType       = NetTypeClient | NetTypeServer
	AnyNetwork       = NetworkTcp | NetworkUdp | NetworkWss
	NetworkTcpClient = NetworkTcp | NetTypeClient
	NetworkTcpServer = NetworkTcp | NetTypeServer
	NetworkUdpClient = NetworkUdp | NetTypeClient
	NetworkUdpServer = NetworkUdp | NetTypeServer
	NetworkWssClient = NetworkWss | NetTypeClient
	NetworkWssServer = NetworkWss | NetTypeServer
	//NetAllowClientReconnect = NetAuthorized | NetTypeClient
	//NetAllowServerReconnect = NetAuthorized | NetTypeServer
)

//Has 判断是否全包含tar
func (this NetType) Has(tar NetType) bool {
	return this&tar == tar
}

//Any 包含中任意一个状态
func (this NetType) Any(tar NetType) bool {
	return this&tar > 0
}

func (this NetType) String() string {
	if this.Has(NetworkTcp) {
		return "tcp"
	} else if this.Has(NetworkUdp) {
		return "udp"
	} else if this.Has(NetworkWss) {
		return "ws"
	}
	return ""
}
