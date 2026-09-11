package session

import (
	"sync"
	"testing"
)

// countingStorage 计数写穿调用,观察 Submit 清脏后 Release 不重复写
type countingStorage struct {
	Storage
	mu     sync.Mutex
	writes int
}

func (f *countingStorage) Update(p *Data, data map[string]any) error {
	f.mu.Lock()
	f.writes++
	f.mu.Unlock()
	return f.Storage.Update(p, data)
}

// TestSubmitClearsDirty Submit 成功后必须清空 dirty:随后的 Release 不得
// 把同一批键重复写进存储(网关的秘钥/uid 写穿路径 Submit 之后,HTTP 框架
// 会在请求结束再调一次 Release)。
func TestSubmitClearsDirty(t *testing.T) {
	cs := &countingStorage{Storage: NewMemory(64)}
	old := Options.Storage
	Options.Storage = cs
	defer func() { Options.Storage = old }()

	ss := NewWithValues("guid-submit", map[string]any{"k": "v"})
	if _, err := ss.New(ss.Data); err != nil {
		t.Fatalf("new error:%v", err)
	}
	if err := ss.Submit(); err != nil {
		t.Fatalf("submit error:%v", err)
	}
	cs.mu.Lock()
	first := cs.writes
	cs.mu.Unlock()
	if first != 1 {
		t.Fatalf("Submit 应写穿一次,实际 %d 次", first)
	}

	ss.Release() //清脏之后:不得重复写

	cs.mu.Lock()
	defer cs.mu.Unlock()
	if cs.writes != first {
		t.Fatalf("Submit 后 Release 重复写入: %d -> %d", first, cs.writes)
	}
}
