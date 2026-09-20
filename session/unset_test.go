package session

import (
	"strings"
	"sync"
	"testing"

	"github.com/hwcer/cosgo/values"
)

// unsetStorage 记录 Unset 调用的包装存储(Unset 已收编为 Storage 强制方法)
type unsetStorage struct {
	Storage
	mu      sync.Mutex
	deleted [][]string
}

func (d *unsetStorage) Unset(p *Data, keys ...string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.deleted = append(d.deleted, keys)
	return nil
}

// spyUpdate 记录 Update 调用的存储,用于断言删除不再走 Update 降级
type spyUpdate struct {
	Storage
	mu      sync.Mutex
	updates []map[string]any
}

func (p *spyUpdate) Update(data *Data, value map[string]any) error {
	cp := make(map[string]any, len(value))
	for k, v := range value {
		cp[k] = v
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.updates = append(p.updates, cp)
	return nil
}

// 🔴 回归:Session.Unset 必须经 Submit 写穿存储(真删除),旧版本 Data.Delete
// 只删内存,Redis 后端下次 Verify 还原副本会读回旧值——删除等于没发生过
func TestSessionUnsetWritesThrough(t *testing.T) {
	old := Options.Storage
	defer func() { Options.Storage = old }()

	us := &unsetStorage{Storage: NewMemory(16)}
	Options.Storage = us

	ss := NewWithValues("u1", values.Values{"gold": 100, "vip": 1})
	if _, err := ss.New(ss.Data); err != nil {
		t.Fatalf("new: %v", err)
	}

	ss.Unset("gold", "vip")
	if err := ss.Submit(); err != nil {
		t.Fatalf("submit: %v", err)
	}
	if len(us.deleted) == 0 {
		t.Fatal("Storage.Unset 应被调用")
	}
	joined := strings.Join(us.deleted[len(us.deleted)-1], ",")
	if !strings.Contains(joined, "gold") || !strings.Contains(joined, "vip") {
		t.Fatalf("删除键应为 gold,vip 实际: %v", us.deleted)
	}
}

// 🔴 Unset 是 Storage 强制方法,不再有"未实现则退化为 Update 写空串"的降级路径:
// 删除只走 Unset,不得把删除键转成 dirty 再写一遍(旧的空串覆盖会污染 Has 语义)
func TestSessionUnsetNoFallbackToUpdate(t *testing.T) {
	old := Options.Storage
	defer func() { Options.Storage = old }()

	sp := &spyUpdate{Storage: NewMemory(16)}
	Options.Storage = sp

	ss := NewWithValues("u2", values.Values{"gold": 100})
	if _, err := ss.New(ss.Data); err != nil {
		t.Fatalf("new: %v", err)
	}

	ss.Unset("gold")
	if err := ss.Submit(); err != nil {
		t.Fatalf("submit: %v", err)
	}
	//删除键不得经 Update 写回(空串或 nil 均不允许)
	for _, u := range sp.updates {
		if _, ok := u["gold"]; ok {
			t.Fatalf("删除键不应经 Update 写回: %v", u)
		}
	}
	//Set 与 Unset 同请求同键:删除优先,Set 的脏标记被 Unset 覆盖后不得回写
	ss2 := NewWithValues("u3", values.Values{"gold": 100})
	if _, err := ss2.New(ss2.Data); err != nil {
		t.Fatalf("new: %v", err)
	}
	ss2.Set("gold", 200)
	ss2.Unset("gold")
	if err := ss2.Submit(); err != nil {
		t.Fatalf("submit: %v", err)
	}
	for _, u := range sp.updates {
		if _, ok := u["gold"]; ok {
			t.Fatalf("同请求 Set 后 Unset,删除键不应经 Update 写回: %v", u)
		}
	}
}

// 🔴 回归:EventSessionCreated 事件参数必须与 events.go 文档一致是 *Data
// (旧实现传入 map,按文档写监听器会静默失效或 panic)
func TestEventSessionCreatedEmitsData(t *testing.T) {
	old := Options.Storage
	defer func() { Options.Storage = old }()
	Options.Storage = NewMemory(16)

	ch := make(chan any, 1)
	On(EventSessionCreated, func(i any) { ch <- i })
	defer On(EventSessionCreated, func(i any) {}) //占位防影响后续用例(无注销API,仅此一次)

	ss := &Session{}
	if _, err := ss.Create("u3", values.Values{"k": 1}); err != nil {
		t.Fatalf("create: %v", err)
	}
	select {
	case i := <-ch:
		if _, ok := i.(*Data); !ok {
			t.Errorf("EventSessionCreated 参数应为 *Data,实际 %T", i)
		}
	default:
		t.Fatal("EventSessionCreated 未触发")
	}
}
