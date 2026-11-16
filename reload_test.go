package cosgo

import (
	"strings"
	"testing"
)

type reloadOrderModule struct {
	log *[]string
}

func (m *reloadOrderModule) Id() string    { return "reload-order" }
func (m *reloadOrderModule) Init() error   { return nil }
func (m *reloadOrderModule) Start() error  { return nil }
func (m *reloadOrderModule) Close() error  { return nil }
func (m *reloadOrderModule) Reload() error { *m.log = append(*m.log, "module"); return nil }

// TestReloadOrder 热重载时序：EventTypReload（开始）→ 各模块 Reload → EventTypReloaded（完成）。
//
// 🔴 两头的语义都有依赖方：yyds options 在 EventTypReload 里重读配置文件、供模块 Reload 用；
// 按新配置做增量注册的业务（如活动排行）在 EventTypReloaded 里读模块重载后的新状态。
func TestReloadOrder(t *testing.T) {
	var log []string
	oldModules, oldReload, oldReloaded := modules, events[EventTypReload], events[EventTypReloaded]
	t.Cleanup(func() {
		modules, events[EventTypReload], events[EventTypReloaded] = oldModules, oldReload, oldReloaded
	})
	modules = []Module{&reloadOrderModule{log: &log}}
	events[EventTypReload] = nil
	events[EventTypReloaded] = nil
	On(EventTypReload, func() error { log = append(log, "reload"); return nil })
	On(EventTypReloaded, func() error { log = append(log, "reloaded"); return nil })

	reload()

	if got := strings.Join(log, ","); got != "reload,module,reloaded" {
		t.Fatalf("热重载时序应为 reload,module,reloaded，实际 %v", got)
	}
	if EventTypReloaded.String() != "EventTypReloaded" {
		t.Fatalf("事件名: %v", EventTypReloaded.String())
	}
}
