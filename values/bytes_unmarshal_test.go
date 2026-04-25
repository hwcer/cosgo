package values

import (
	"testing"
)

// TestBytes_UnmarshalJSON_NullKeepsEmpty: JSON null 不应覆盖接收器。
// 修复前条件反了,首次 Unmarshal 任何输入都会被写入。
func TestBytes_UnmarshalJSON_NullKeepsEmpty(t *testing.T) {
	var b Bytes
	if err := b.UnmarshalJSON([]byte("null")); err != nil {
		t.Fatalf("unmarshal null: %v", err)
	}
	if len(b) != 0 {
		t.Errorf("expected empty after null, got %q", b)
	}
}

func TestBytes_UnmarshalJSON_NonNullStored(t *testing.T) {
	var b Bytes
	in := []byte(`{"k":1}`)
	if err := b.UnmarshalJSON(in); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if string(b) != string(in) {
		t.Errorf("got %q, want %q", b, in)
	}
}

func TestBytes_UnmarshalJSON_NullAfterValue(t *testing.T) {
	// 已有值再遇到 null 应保留原值(按 MarshalJSON 的 null 语义对称)
	b := Bytes([]byte(`"x"`))
	if err := b.UnmarshalJSON([]byte("null")); err != nil {
		t.Fatalf("unmarshal null: %v", err)
	}
	if string(b) != `"x"` {
		t.Errorf("null should not overwrite, got %q", b)
	}
}
