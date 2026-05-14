package values

import (
	"encoding/json"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type TestEquipProperty struct {
	Id  int32 `bson:"id" json:"id"`
	Val int64 `bson:"val" json:"val"`
}

// TestValuesMapCompatibility 测试 Values 与 Map 的兼容性：
// Values 存对象 -> BSON 序列化 -> Map 反序列化 -> Map.Unmarshal 验证
func TestValuesMapCompatibility(t *testing.T) {
	type PlayerInfo struct {
		Name  string `bson:"name" json:"name"`
		Level int32  `bson:"level" json:"level"`
		Score int64  `bson:"score" json:"score"`
	}

	// 1. 使用 Values 储存一个对象
	vs := make(Values)
	info := PlayerInfo{Name: "alice", Level: 10, Score: 9999}
	vs.Set("info", info)

	props := []TestEquipProperty{
		{Id: 1, Val: 100},
		{Id: 2, Val: 200},
	}
	vs.Set("props", props)
	vs.Set("gold", int64(12345))

	t.Logf("Values original: info=%+v, props=%+v, gold=%v", vs["info"], vs["props"], vs["gold"])

	// 2. 使用 BSON 序列化
	raw, err := bson.Marshal(vs)
	if err != nil {
		t.Fatalf("bson.Marshal(Values): %v", err)
	}
	t.Logf("BSON bytes len: %d", len(raw))

	// 3. 使用 Map 反序列化二进制
	m := make(Attach[string])
	if err = bson.Unmarshal(raw, &m); err != nil {
		t.Fatalf("bson.Unmarshal -> Map: %v", err)
	}
	t.Logf("Map keys after unmarshal: info(%T), props(%T), gold(%T)", m["info"], m["props"], m["gold"])

	// 4. 使用 Map.Unmarshal 观察值是不是正确

	// 4a. 结构体
	var gotInfo PlayerInfo
	if err = m.Unmarshal("info", &gotInfo); err != nil {
		t.Fatalf("Map.Unmarshal info: %v", err)
	}
	if gotInfo.Name != "alice" || gotInfo.Level != 10 || gotInfo.Score != 9999 {
		t.Fatalf("info mismatch: got %+v, want {alice 10 9999}", gotInfo)
	}
	t.Logf("info OK: %+v", gotInfo)

	// 4b. 切片
	var gotProps []TestEquipProperty
	if err = m.Unmarshal("props", &gotProps); err != nil {
		t.Fatalf("Map.Unmarshal props: %v", err)
	}
	if len(gotProps) != 2 {
		t.Fatalf("props len: got %d, want 2", len(gotProps))
	}
	if gotProps[0].Id != 1 || gotProps[0].Val != 100 {
		t.Fatalf("props[0] mismatch: %+v", gotProps[0])
	}
	if gotProps[1].Id != 2 || gotProps[1].Val != 200 {
		t.Fatalf("props[1] mismatch: %+v", gotProps[1])
	}
	t.Logf("props OK: %+v", gotProps)

	// 4c. 基础类型
	gold := m.GetInt64("gold")
	if gold != 12345 {
		t.Fatalf("gold mismatch: got %d, want 12345", gold)
	}
	t.Logf("gold OK: %d", gold)
}

func TestMapUnmarshalRoundTrip(t *testing.T) {
	type AttachKey int32
	const keyProps AttachKey = 6003

	// 1. 写入 Go 对象
	m := make(Attach[AttachKey])
	props := []TestEquipProperty{
		{Id: 1, Val: 100},
		{Id: 2, Val: 200},
	}
	m.Set(keyProps, props)

	// 2. BSON 序列化（模拟写入 MongoDB）
	doc := bson.D{{Key: "attach", Value: map[string]any{
		"6003": m[keyProps],
	}}}
	raw, err := bson.Marshal(doc)
	if err != nil {
		t.Fatalf("bson.Marshal: %v", err)
	}
	t.Logf("BSON bytes len: %d", len(raw))

	// 3. 从 BSON 字节加载（模拟从 MongoDB 读取）
	var loaded struct {
		Attach map[string]any `bson:"attach"`
	}
	if err = bson.Unmarshal(raw, &loaded); err != nil {
		t.Fatalf("bson.Unmarshal: %v", err)
	}

	m2 := make(Attach[AttachKey])
	for k, v := range loaded.Attach {
		var ki int32
		if _, err := fmt.Sscanf(k, "%d", &ki); err == nil {
			m2[AttachKey(ki)] = v
		}
	}

	// 验证 DB 读出的值是 bson.A（不是原始 Go 类型）
	raw6003 := m2[keyProps]
	if _, ok := raw6003.(bson.A); !ok {
		t.Fatalf("expected bson.A from DB, got %T", raw6003)
	}
	t.Logf("DB value type: %T", raw6003)

	// 3.5 bson.A 二次 BSON 序列化：模拟再次写入 DB
	doc2 := bson.D{{Key: "attach", Value: map[string]any{
		"6003": m2[keyProps],
	}}}
	raw2, err := bson.Marshal(doc2)
	if err != nil {
		t.Fatalf("second bson.Marshal: %v", err)
	}
	t.Logf("second BSON bytes len: %d", len(raw2))

	// 从二次 BSON 中加载验证
	var loaded2 struct {
		Attach map[string]any `bson:"attach"`
	}
	if err = bson.Unmarshal(raw2, &loaded2); err != nil {
		t.Fatalf("second bson.Unmarshal: %v", err)
	}
	if _, ok := loaded2.Attach["6003"].(bson.A); !ok {
		t.Fatalf("second load: expected bson.A, got %T", loaded2.Attach["6003"])
	}
	t.Logf("second BSON round-trip OK, type: %T", loaded2.Attach["6003"])

	// 3.6 bson.A 状态下 MarshalJSON
	jsonBefore, err := m2.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON before Unmarshal: %v", err)
	}
	t.Logf("JSON (bson.A): %s", jsonBefore)

	var checkMap map[string]json.RawMessage
	if err = json.Unmarshal(jsonBefore, &checkMap); err != nil {
		t.Fatalf("json parse bson.A output: %v", err)
	}
	if _, ok := checkMap["6003"]; !ok {
		t.Fatal("JSON (bson.A) missing key 6003")
	}

	// 4. Unmarshal 到具体类型
	var result []TestEquipProperty
	if err = m2.Unmarshal(keyProps, &result); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}
	if result[0].Id != 1 || result[0].Val != 100 {
		t.Fatalf("item[0] mismatch: %+v", result[0])
	}
	if result[1].Id != 2 || result[1].Val != 200 {
		t.Fatalf("item[1] mismatch: %+v", result[1])
	}
	t.Logf("Unmarshal result: %+v", result)

	// 5. 验证缓存：再次 Unmarshal 应走 reflect 快速路径
	cached := m2[keyProps]
	if _, ok := cached.([]TestEquipProperty); !ok {
		t.Fatalf("expected cached Go type after first Unmarshal, got %T", cached)
	}

	var result2 []TestEquipProperty
	if err = m2.Unmarshal(keyProps, &result2); err != nil {
		t.Fatalf("second Unmarshal: %v", err)
	}
	if len(result2) != 2 || result2[0].Id != 1 || result2[1].Val != 200 {
		t.Fatalf("cached Unmarshal mismatch: %+v", result2)
	}

	// 6. 验证深拷贝：修改 result2 不影响缓存
	result2[0].Val = 999
	var result3 []TestEquipProperty
	if err = m2.Unmarshal(keyProps, &result3); err != nil {
		t.Fatalf("third Unmarshal: %v", err)
	}
	if result3[0].Val != 100 {
		t.Fatalf("deep copy broken: expected 100, got %d", result3[0].Val)
	}
	t.Log("deep copy verified: mutation did not affect cache")

	// 7. MarshalJSON
	jsonBytes, err := m2.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	t.Logf("JSON: %s", jsonBytes)

	var jsonMap map[string]json.RawMessage
	if err = json.Unmarshal(jsonBytes, &jsonMap); err != nil {
		t.Fatalf("json parse: %v", err)
	}
	if _, ok := jsonMap["6003"]; !ok {
		t.Fatal("JSON missing key 6003")
	}
}
