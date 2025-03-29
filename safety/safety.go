package safety

import (
	"github.com/hwcer/cosgo/utils"
	"strings"
	"sync"
)

type ReloadHandle func(int) (name string, rule string, status Status, ok bool)

/*
10.0.0.0~10.255.255.255（A类）
172.16.0.0~172.31.255.255（B类）
192.168.0.0~192.168.255.255（C类）
*/
//内网IP段
var localAddress = []string{"127.0.0.1", "10.0.0.0~10.255.255.255", "172.16.0.0~172.31.255.255", "192.168.0.0~192.168.255.255"}

func New() *Safety {
	return &Safety{rules: NewSafetyData()}
}

// SafetyUpdate 更新器 返回true 替换原来的规则
type SafetyUpdate func(*SafetyData) bool

// SafetyRule IP匹配规则
type SafetyRule struct {
	local    bool      // 是否内网地址
	status   Status    //状态
	matching [2]uint64 //匹配范围
}

func (this *SafetyRule) Parse(rule string, status Status, local bool) {
	this.local = local
	this.status = status
	this.matching = [2]uint64{0, 0}
	arr := strings.Split(rule, "~")
	this.matching[0] = utils.IPv4Encode(arr[0])
	if len(arr) > 1 {
		this.matching[1] = utils.IPv4Encode(arr[1])
	}
}

func (this *SafetyRule) Match(ips uint64, useLocalAddress bool) bool {
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

func NewSafetyData() *SafetyData {
	return &SafetyData{dict: make(map[string]*SafetyRule)}
}

type SafetyData struct {
	dict map[string]*SafetyRule
}

func (this *SafetyData) Get(k string) *SafetyRule {
	return this.dict[k]
}
func (this *SafetyData) Copy() *SafetyData {
	d := NewSafetyData()
	for k, v := range this.dict {
		d.dict[k] = v
	}
	return d
}

func (this *SafetyData) Delete(id string) *SafetyData {
	d := NewSafetyData()
	for k, v := range this.dict {
		if k != id {
			d.dict[k] = v
		}
	}
	return d
}

// Setter 仅仅 在 Safety.Update中执行
func (this *SafetyData) Setter(name string, rule string, status Status, local bool) {
	r := &SafetyRule{}
	r.Parse(rule, status, local)
	this.dict[name] = r
}

type Safety struct {
	rules           *SafetyData
	mutex           sync.Mutex
	useLocalAddress bool //是否使用内网IP(内网IP一律通过)
}

func (this *Safety) Lock(f func()) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	f()
}

func (this *Safety) Get(name string) *SafetyRule {
	return this.rules.Get(name)
}

// Reload 加载所有规则
func (this *Safety) Reload(f SafetyUpdate) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	rules := this.rules.Copy()
	if f(rules) {
		this.rules = rules
	}
}

func (this *Safety) Update(name string, rule string, status Status, local bool) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	rules := this.rules.Copy()
	rules.Setter(name, rule, status, local)
	this.rules = rules
}

// Delete 删除规则
func (this *Safety) Delete(name string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.rules = this.rules.Delete(name)
}

func (this *Safety) Match(ip string, useLocalAddress bool) Status {
	ips := utils.IPv4Encode(ip)
	if ips == 0 {
		return StatusNone
	}
	for _, v := range this.rules.dict {
		if v.Match(ips, useLocalAddress) {
			return v.status
		}
	}
	return StatusNone
}

// UseLocalAddress 信任所有内网IP
func (this *Safety) UseLocalAddress() {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	rules := this.rules.Copy()
	for _, s := range localAddress {
		rules.Setter(s, s, StatusEnable, true)
	}
	this.rules = rules
	this.useLocalAddress = true
}
