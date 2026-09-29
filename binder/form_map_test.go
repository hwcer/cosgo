package binder

import (
	"net/url"
	"testing"
)

// form 反序列化进 map:键必须 string,值 string 或 interface 皆可。
// 回归背景:5f19195 的防御校验一度要求 map[string]string,误伤了 cosweb
// Body 参数通道的 values.Values(map[string]any)——form 表单参数全丢。
func TestFormUnmarshalIntoMap(t *testing.T) {
	vs := url.Values{}
	vs.Set("id", "11614brkd11")
	vs.Set("name", "微信小游戏")
	vs.Set("empty", "")

	f := &formBinding{}

	// map[string]any(interface 元素):合法,值以 string 落入
	anyMap := map[string]any{}
	if err := f.UnmarshalFromValues(vs, &anyMap); err != nil {
		t.Fatalf("map[string]any should be accepted: %v", err)
	}
	if anyMap["id"] != "11614brkd11" || anyMap["name"] != "微信小游戏" {
		t.Fatalf("map[string]any content = %v", anyMap)
	}
	if _, ok := anyMap["empty"]; ok {
		t.Fatal("empty value should be skipped")
	}

	// map[string]string:原本就支持
	strMap := map[string]string{}
	if err := f.UnmarshalFromValues(vs, &strMap); err != nil {
		t.Fatalf("map[string]string should be accepted: %v", err)
	}
	if strMap["id"] != "11614brkd11" {
		t.Fatalf("map[string]string content = %v", strMap)
	}

	// nil map 自动初始化
	var nilMap map[string]any
	if err := f.UnmarshalFromValues(vs, &nilMap); err != nil {
		t.Fatalf("nil map[string]any should be accepted: %v", err)
	}
	if nilMap["id"] != "11614brkd11" {
		t.Fatalf("nil map content = %v", nilMap)
	}

	// 非 string 键 / 非 string 非 interface 值:必须拒绝(防 SetMapIndex panic)
	intMap := map[string]int{}
	if err := f.UnmarshalFromValues(vs, &intMap); err == nil {
		t.Fatal("map[string]int should be rejected")
	}
	nonStrKey := map[int]string{}
	if err := f.UnmarshalFromValues(vs, &nonStrKey); err == nil {
		t.Fatal("map[int]string should be rejected")
	}
}
