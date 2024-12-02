package options

import (
	"fmt"
	"strconv"
)

const (
	ServiceSelectorAverage       = "_rpc_srv_avg"
	ServiceSelectorServerId      = "_rpc_srv_sid"  //服务器编号
	ServiceSelectorServerAddress = "_rpc_srv_addr" //rpc服务器ID,selector 中固定转发地址

	ServiceMetadataUID       = "uid"
	ServiceMetadataGUID      = "guid"
	ServiceMetadataServerId  = "sid"
	ServiceMetadataRequestId = "_rid"

	ServiceMessagePath   = "_msg_path"
	ServiceMessageRoom   = "_msg_room"
	ServiceMessageIgnore = "_msg_ignore"

	ServicePlayerOAuth  = "_player_oauth"
	ServicePlayerLogout = "_player_logout"

	ServicePlayerRoomJoin  = "player.room.join"  //已经加入的房间
	ServicePlayerRoomLeave = "player.room.leave" //离开房间
	ServicePlayerSelector  = "service.selector." //服务器重定向
)

// NewMetadata 创建新Metadata，参数k1,v1,k2,v2...
func NewMetadata(args ...string) Metadata {
	r := Metadata{}
	var i, j int
	for i = 0; i < len(args)-1; i += 2 {
		j = i + 1
		r[args[i]] = args[j]
	}
	return r
}

type Metadata map[string]string

func (this Metadata) Set(k string, v any) {
	this[k] = fmt.Sprintf("%v", v)
}

func (this Metadata) SetAddress(v string) {
	this[ServiceSelectorServerAddress] = v
}

func (this Metadata) SetServerId(v int32) {
	this[ServiceSelectorServerId] = strconv.Itoa(int(v))
}
func (this Metadata) SetContentType(v string) {
	this["Content-Type"] = v
}

func (this Metadata) Json() map[string]string {
	return this
}
