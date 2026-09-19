package session

import (
	"strings"
	"sync"
	"testing"

	"github.com/hwcer/cosgo/values"
)

// deleterStorage 记录 DeleteKeys 调用的包装存储
type deleterStorage struct {
	Storage
	mu    sync.Mutex
	deleted [][]string
}

func (d *deleterStorage) DeleteKeys(p *Data, keys ...string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.deleted = append(d.deleted, keys)
	return nil
}

// plainStorage 不实现 StorageDeleter:Unset 应退化为 Update 写空串
type plainStorage struct {
	Storage
	mu     sync.Mutex
	updates []map[string]any
}

func (p *plainStorage) Update(data *Data, value map[string]any) error {
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

	ds := &deleterStorage{Storage: NewMemory(16)}
	Options.Storage = ds

	ss := NewWithValues("u1", values.Values{"gold": 100, "vip": 1})
	if _, err := ss.New(ss.Data); err != nil {
		t.Fatalf("new: %v", err)
	}

	ss.Unset("gold", "vip")
	if err := ss.Submit(); err != nil {
		t.Fatalf("submit: %v", err)
	}
	if len(ds.deleted) == 0 {
		t.Fatal("StorageDeleter.DeleteKeys 应被调用")
	}
	joined := strings.Join(ds.deleted[len(ds.deleted)-1], ",")
	if !strings.Contains(joined, "gold") || !strings.Contains(joined, "vip") {
		t.Fatalf("删除键应为 gold,vip 实际: %v", ds.deleted)
	}
}

// 未实现 StorageDeleter 的存储:退化为写空串(与旧行为兼容,不丢语义)
func TestSessionUnsetFallbackWritesEmpty(t *testing.T) {
	old := Options.Storage
	defer func() { Options.Storage = old }()

	ps := &plainStorage{Storage: NewMemory(16)}
	Options.Storage = ps

	ss := NewWithValues("u2", values.Values{"gold": 100})
	if _, err := ss.New(ss.Data); err != nil {
		t.Fatalf("new: %v", err)
	}

	ss.Unset("gold")
	if err := ss.Submit(); err != nil {
		t.Fatalf("submit: %v", err)
	}
	if len(ps.updates) == 0 {
		t.Fatal("退化为 Update 写空串,Update 应被调用")
	}
	last := ps.updates[len(ps.updates)-1]
	if v, ok := last["gold"]; !ok || v != nil {
		t.Fatalf("删除键应以 nil(空串)写入: %v", last)
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
