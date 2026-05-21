package values

import (
	"encoding/json"
	"testing"
)

// TestBytes_UnmarshalJSON_NullKeepsEmpty: JSON null 不应覆盖接收器。
func TestBytes_UnmarshalJSON_NullKeepsEmpty(t *testing.T) {
	var b Bytes
	if err := b.UnmarshalJSON([]byte("null")); err != nil {
		t.Fatalf("unmarshal null: %v", err)
	}
	if len(b) != 0 {
		t.Errorf("expected empty after null, got %q", b)
	}
}

// TestBytes_JSON_Roundtrip: JSON→BSON→JSON 往返一致
func TestBytes_JSON_Roundtrip(t *testing.T) {
	in := `{"k":1}`
	var b Bytes
	if err := b.UnmarshalJSON([]byte(in)); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty after unmarshal")
	}
	out, err := b.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var want, got map[string]any
	json.Unmarshal([]byte(in), &want)
	json.Unmarshal(out, &got)
	if got["k"] != want["k"] {
		t.Errorf("roundtrip mismatch: got %v, want %v", got, want)
	}
}

// TestBytes_BSON_Roundtrip: Marshal→Unmarshal 往返一致
func TestBytes_BSON_Roundtrip(t *testing.T) {
	type Obj struct {
		Name string `bson:"name"`
		Val  int32  `bson:"val"`
	}
	src := Obj{Name: "test", Val: 42}
	var b Bytes
	if err := b.Marshal(src); err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var dst Obj
	if err := b.Unmarshal(&dst); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if dst != src {
		t.Errorf("got %+v, want %+v", dst, src)
	}
}

func TestBytes_UnmarshalJSON_NullAfterValue(t *testing.T) {
	var b Bytes
	b.UnmarshalJSON([]byte(`{"x":1}`))
	prev := make([]byte, len(b))
	copy(prev, b)
	if err := b.UnmarshalJSON([]byte("null")); err != nil {
		t.Fatalf("unmarshal null: %v", err)
	}
	if string(b) != string(prev) {
		t.Errorf("null should not overwrite")
	}
}
