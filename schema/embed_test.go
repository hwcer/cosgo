package schema

import (
	"testing"
)

type EmbedInner struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type embedOuter struct {
	ID int `json:"id"`
	EmbedInner
}

// TestEmbeddedFieldAfterWarmup 验证匿名嵌入字段被复制到外层 Fields 后,
// getValueFn 指向正确的 Index。若 warmup 在内层先编译了 fn,复制到外层时
// 如果不清空 fn 指针,会读到错误字段。
func TestEmbeddedFieldAfterWarmup(t *testing.T) {
	s, err := Parse(&embedOuter{})
	if err != nil {
		t.Fatal(err)
	}
	v := &embedOuter{ID: 7, EmbedInner: EmbedInner{Name: "alice", Age: 30}}

	nameField := s.LookUpField("Name")
	if nameField == nil {
		t.Fatal("Name field not found")
	}
	got := nameField.Get(ValueOf(v)).Interface()
	if got != "alice" {
		t.Errorf("promoted field Name via outer schema: got %v, want alice", got)
	}

	ageField := s.LookUpField("Age")
	if ageField == nil {
		t.Fatal("Age field not found")
	}
	gotAge := ageField.Get(ValueOf(v)).Interface()
	if gotAge != 30 {
		t.Errorf("promoted field Age via outer schema: got %v, want 30", gotAge)
	}

	idField := s.LookUpField("ID")
	if idField == nil {
		t.Fatal("ID field not found")
	}
	gotID := idField.Get(ValueOf(v)).Interface()
	if gotID != 7 {
		t.Errorf("direct field ID: got %v, want 7", gotID)
	}
}
