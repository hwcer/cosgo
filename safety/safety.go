package safety

import (
	"github.com/hwcer/cosgo/utils"
	"sync"
)

func New() *Safety {
	return &Safety{dict: make(map[string]*Rule)}
}

type ReloadHandle func(int) (name string, rule string, status Status, ok bool)

type Safety struct {
	dict            map[string]*Rule
	mutex           sync.Mutex
	useLocalAddress bool
}

func (this *Safety) copy() map[string]*Rule {
	r := make(map[string]*Rule)
	for k, v := range this.dict {
		r[k] = v
	}
	return r
}

func (this *Safety) lock(f func()) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	f()
}

func (this *Safety) Get(name string) *Rule {
	return this.dict[name]
}

func (this *Safety) Reload(length int, f ReloadHandle) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	rules := make(map[string]*Rule)
	for i := 0; i < length; i++ {
		if name, rule, status, ok := f(i); ok {
			r := &Rule{}
			r.Parse(rule, status, false)
			rules[name] = r
		}
	}
	if this.useLocalAddress {
		for _, s := range localAddress {
			r := &Rule{}
			r.Parse(s, StatusEnable, true)
			rules[s] = r
		}
	}
	this.dict = rules
}

//Create 创建规则
func (this *Safety) Create(name string, rule string, status Status) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	rules := this.copy()
	r := &Rule{}
	r.Parse(rule, status, false)
	rules[name] = r
	this.dict = rules
}

//Delete 删除规则
func (this *Safety) Delete(name string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	rules := this.copy()
	delete(rules, name)
	this.dict = rules
}

func (this *Safety) Match(ip string, useLocalAddress bool) Status {
	ips := utils.Ipv4Encode(ip)
	if ips == 0 {
		return StatusNone
	}
	for _, v := range this.dict {
		if v.Match(ips, useLocalAddress) {
			return v.status
		}
	}
	return StatusNone
}

//UseLocalAddress 信任所有内网IP
func (this *Safety) UseLocalAddress() {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	rules := this.copy()
	for _, s := range localAddress {
		r := &Rule{}
		r.Parse(s, StatusEnable, true)
		rules[s] = r
	}
	this.dict = rules
	this.useLocalAddress = true
}
