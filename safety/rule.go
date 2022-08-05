package safety

import (
	"github.com/hwcer/cosgo/utils"
	"strings"
)

//Rule IP匹配规则
type Rule struct {
	local    bool      // 是否内网地址
	status   Status    //状态
	matching [2]uint32 //匹配范围
}

func (this *Rule) Parse(rule string, status Status, local bool) {
	this.local = local
	this.status = status
	this.matching = [2]uint32{0, 0}
	arr := strings.Split(rule, "~")
	this.matching[0] = utils.Ipv4Encode(arr[0])
	if len(arr) > 1 {
		this.matching[1] = utils.Ipv4Encode(arr[1])
	}
}

func (this *Rule) Match(ips uint32, useLocalAddress bool) bool {
	if this.local && !useLocalAddress {
		return false
	}
	if this.matching[1] == 0 && ips == this.matching[0] {
		return true
	}
	if this.matching[1] > 0 && ips >= this.matching[0] && ips <= this.matching[1] {
		return true
	}
	return false
}
