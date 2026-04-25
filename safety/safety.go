package safety

import (
	"sync"
	"sync/atomic"
)

// 内网 IP 段
var localAddress = []string{
	"127.0.0.1",
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
}

// ReloadHandle 规则加载回调（遗留兼容）
type ReloadHandle func(int) (name string, rule string, status Status, ok bool)

// SafetyUpdate 更新器回调，返回 true 表示应用变更
type SafetyUpdate func(*SafetyData) bool

// SafetyRule IP 匹配规则
type SafetyRule struct {
	name   string    // 规则名称
	local  bool      // 是否内网地址规则
	status Status    // 白名单 / 黑名单
	start  uint32    // IP 范围起始（含）
	end    uint32    // IP 范围结束（含），等于 start 时为精确匹配
}

// Match 检查 ip 是否命中本规则
func (r *SafetyRule) Match(ip uint32, useLocalAddress bool) bool {
	if r.local && !useLocalAddress {
		return false
	}
	return ip >= r.start && ip <= r.end
}

// SafetyData 不可变规则集（CoW 语义）
// dict 用于按名查找，list 用于 Match 遍历（slice 迭代比 map 快且顺序稳定）
type SafetyData struct {
	dict map[string]*SafetyRule
	list []*SafetyRule
}

func NewSafetyData() *SafetyData {
	return &SafetyData{dict: make(map[string]*SafetyRule)}
}

// Get 按名查找规则
func (d *SafetyData) Get(k string) *SafetyRule {
	return d.dict[k]
}

// Copy 浅拷贝（CoW 写时复制的"写"阶段调用）
func (d *SafetyData) Copy() *SafetyData {
	n := &SafetyData{
		dict: make(map[string]*SafetyRule, len(d.dict)),
		list: make([]*SafetyRule, 0, len(d.list)),
	}
	for k, v := range d.dict {
		n.dict[k] = v
	}
	n.list = append(n.list, d.list...)
	return n
}

// Delete 返回移除指定规则后的新快照
func (d *SafetyData) Delete(name string) *SafetyData {
	if _, ok := d.dict[name]; !ok {
		return d
	}
	n := &SafetyData{
		dict: make(map[string]*SafetyRule, len(d.dict)),
		list: make([]*SafetyRule, 0, len(d.list)),
	}
	for k, v := range d.dict {
		if k != name {
			n.dict[k] = v
			n.list = append(n.list, v)
		}
	}
	return n
}

// Setter 添加或更新规则（在 Safety.mutex 下调用）
func (d *SafetyData) Setter(name string, rule string, status Status, local bool) {
	start, end := parseRule(rule)
	r := &SafetyRule{
		name:   name,
		local:  local,
		status: status,
		start:  start,
		end:    end,
	}
	if old, ok := d.dict[name]; ok {
		// 替换 list 中的旧规则
		for i, v := range d.list {
			if v == old {
				d.list[i] = r
				break
			}
		}
	} else {
		d.list = append(d.list, r)
	}
	d.dict[name] = r
}

// New 创建 Safety 实例
func New() *Safety {
	s := &Safety{}
	s.rules.Store(NewSafetyData())
	return s
}

// Safety 网络黑白名单检查器
//
// 并发模型：
//   - 读路径（Match/Get）通过 atomic.Pointer 无锁读取当前规则快照
//   - 写路径（Update/Delete/Reload）使用 mutex 串行化 CoW 更新
//   - Match 使用零分配 IPv4 内联解析 + slice 顺序遍历
type Safety struct {
	rules           atomic.Pointer[SafetyData]
	mutex           sync.Mutex
	useLocalAddress bool
}

// loadRules 原子读取当前规则快照
func (s *Safety) loadRules() *SafetyData {
	return s.rules.Load()
}

// Lock 在 mutex 下执行函数
func (s *Safety) Lock(f func()) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	f()
}

// Get 按名查找规则（无锁）
func (s *Safety) Get(name string) *SafetyRule {
	return s.loadRules().Get(name)
}

// Reload 批量更新规则
func (s *Safety) Reload(f SafetyUpdate) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	rules := s.loadRules().Copy()
	if f(rules) {
		s.rules.Store(rules)
	}
}

// Update 添加或更新单条规则
func (s *Safety) Update(name string, rule string, status Status, local bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	rules := s.loadRules().Copy()
	rules.Setter(name, rule, status, local)
	s.rules.Store(rules)
}

// Delete 删除规则
func (s *Safety) Delete(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.rules.Store(s.loadRules().Delete(name))
}

// Match 检查 IP 是否命中任一规则，返回命中规则的状态
// 零分配内联 IPv4 解析 + slice 顺序遍历，热路径无锁
func (s *Safety) Match(ip string, useLocalAddress bool) Status {
	ipVal := parseIPv4(ip)
	if ipVal == 0 {
		return StatusNone
	}
	data := s.loadRules()
	for _, r := range data.list {
		if r.Match(ipVal, useLocalAddress) {
			return r.status
		}
	}
	return StatusNone
}

// UseLocalAddress 信任所有内网 IP（127.0.0.1 + 三大私网段）
func (s *Safety) UseLocalAddress() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	rules := s.loadRules().Copy()
	for _, addr := range localAddress {
		rules.Setter(addr, addr, StatusEnable, true)
	}
	s.rules.Store(rules)
	s.useLocalAddress = true
}
