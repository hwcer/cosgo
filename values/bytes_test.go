package values

import (
	"encoding/json"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type bytesTestStruct struct {
	Name string `bson:"name" json:"name"`
	Val  int32  `bson:"val" json:"val"`
}

// bytesTestDoc 用于测试 Bytes 作为 BSON 字段的序列化
type bytesTestDoc struct {
	Data Bytes `bson:"data" json:"data"`
}

// bytesTestCase 每种数据类型的测试用例
type bytesTestCase struct {
	name string
	src  any
	dst  any // Unmarshal 目标，类型需与 src 匹配
}

func TestBytes_AllTypes(t *testing.T) {
	cases := []bytesTestCase{
		{name: "int32", src: bson.M{"v": int32(100)}, dst: &bson.M{}},
		{name: "string", src: bson.M{"v": "hello"}, dst: &bson.M{}},
		{name: "map", src: map[string]any{"key": "value", "num": int32(1)}, dst: &map[string]any{}},
		{name: "arr", src: bson.M{"v": bson.A{int32(1), int32(2), int32(3)}}, dst: &bson.M{}},
		{name: "struct", src: bytesTestStruct{Name: "test", Val: 42}, dst: &bytesTestStruct{}},
	}

	for i := range cases {
		tc := &cases[i]
		t.Run(tc.name, func(t *testing.T) {
			// 1. Bytes.Marshal: Go值 → BSON二进制
			var b Bytes
			if err := b.Marshal(tc.src); err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			t.Logf("BSON len: %d", len(b))

			// 2. json.Marshal: BSON二进制 → JSON字符串
			jsonData, err := json.Marshal(&b)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}
			t.Logf("JSON: %s", jsonData)

			// 3. json.Unmarshal: JSON字符串 → BSON二进制（通过 UnmarshalJSON）
			var b2 Bytes
			if err = json.Unmarshal(jsonData, &b2); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			// 验证 JSON 往返后能正确 Unmarshal 回 Go 值
			dst2 := newDst(tc.dst)
			if err = b2.Unmarshal(dst2); err != nil {
				t.Fatalf("Unmarshal after JSON roundtrip: %v", err)
			}
			t.Logf("JSON roundtrip: %+v", deref(dst2))

			// 4. BSON序列化: 模拟存入MongoDB（MarshalBSONValue）再读出（UnmarshalBSONValue）
			doc := bytesTestDoc{Data: b}
			raw, err := bson.Marshal(doc)
			if err != nil {
				t.Fatalf("bson.Marshal doc: %v", err)
			}
			t.Logf("BSON doc len: %d", len(raw))

			var doc2 bytesTestDoc
			if err = bson.Unmarshal(raw, &doc2); err != nil {
				t.Fatalf("bson.Unmarshal doc: %v", err)
			}
			// 验证 BSON 往返后能正确 Unmarshal 回 Go 值
			if err = doc2.Data.Unmarshal(tc.dst); err != nil {
				t.Fatalf("Unmarshal after BSON roundtrip: %v", err)
			}
			t.Logf("BSON roundtrip: %+v", deref(tc.dst))
		})
	}
}

func newDst(v any) any {
	switch v.(type) {
	case *bson.M:
		return &bson.M{}
	case *map[string]any:
		return &map[string]any{}
	case *bytesTestStruct:
		return &bytesTestStruct{}
	}
	return nil
}

func deref(v any) any {
	switch p := v.(type) {
	case *bson.M:
		return *p
	case *map[string]any:
		return *p
	case *bytesTestStruct:
		return *p
	}
	return v
}
