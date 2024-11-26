package options

type protocol int8

const (
	ProtocolTypeWSS  int8 = 1 << 0
	ProtocolTypeTCP  int8 = 1 << 1
	ProtocolTypeHTTP int8 = 1 << 2
)

const (
	ServiceAddressPrefix = "_service_address_"
	ServiceMetadataUID   = "_srv_uid"
	ServiceMetadataGUID  = "_srv_guid"
)

func (p protocol) Has(t int8) bool {
	v := int8(p)
	return v|t == v
}

// CMux 是否启动cmux模块
func (p protocol) CMux() bool {
	if !p.Has(ProtocolTypeTCP) {
		return false
	}
	return p.Has(ProtocolTypeWSS) || p.Has(ProtocolTypeHTTP)
}

var Gate = &gate{
	Prefix:    "handle",
	Address:   "0.0.0.0:80",
	Protocol:  2,
	Broadcast: 1,
	Websocket: "ws",
}

type gate = struct {
	Prefix    string   `json:"prefix"`    //路由强制前缀
	Address   string   `json:"address"`   //连接地址
	Protocol  protocol `json:"protocol"`  //1-短链接，2-长连接，3-长短链接全开
	Broadcast int8     `json:"broadcast"` //Push message 0-关闭，1-双向通信，2-独立启动服务器,推送消息必须启用长链接
	Websocket string   `json:"websocket"` //开启websocket时,路由前缀
	WSVerify  bool     `json:"WSVerify"`
}

//type Route struct {
//	Prefix string `json:"prefix"` //路由强制前缀
//}

//type Metadata struct {
//	API  string `json:"api"`  //socket 推送消息时的路径(协议)
//	UID  string `json:"uid"`  //角色ID
//	GUID string `json:"guid"` //账号ID
//}

//func GetServiceAddress(k string) string {
//	return ServiceAddressPrefix + k
//}
