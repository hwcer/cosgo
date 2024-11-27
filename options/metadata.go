package options

import (
	"fmt"
	"strconv"
)

const (
	ServicesMetadataSelectorAverage       = "_rpc_srv_avg"
	ServicesMetadataSelectorServerId      = "_rpc_srv_sid"  //服务器编号
	ServicesMetadataSelectorServerAddress = "_rpc_srv_addr" //rpc服务器ID,selector 中固定转发地址

	ServiceMetadataApi       = "_api"
	ServiceMetadataUID       = "uid"
	ServiceMetadataGUID      = "guid"
	ServiceMetadataServerId  = "sid"
	ServiceMetadataRequestId = "_rid"
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
	this[ServicesMetadataSelectorServerAddress] = v
}

func (this Metadata) SetServerId(v int32) {
	this[ServicesMetadataSelectorServerId] = strconv.Itoa(int(v))
}
func (this Metadata) SetContentType(v string) {
	this["Content-Type"] = v
}

func (this Metadata) Json() map[string]string {
	return this
}
