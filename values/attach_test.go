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

// TestAttachUnmarshalPointerValueCompat 内存里直接读（热路径）时兼容两种存法：
// Set 原样保存调用方给的对象（指针），冷路径缓存的却是读入目标所指的值类型。
func TestAttachUnmarshalPointerValueCompat(t *testing.T) {
	m := make(Values)
	m.Set("ptr", &TestEquipProperty{Id: 1, Val: 100})
	m.Set("val", TestEquipProperty{Id: 2, Val: 200})
	other := struct{ X int }{}

	// 存指针 → 读入值
	var a TestEquipProperty
	if err := m.Unmarshal("ptr", &a); err != nil || a.Id != 1 || a.Val != 100 {
		t.Fatalf("pointer -> value: got %+v, err %v", a, err)
	}
	// 存值 → 读入指针
	var b *TestEquipProperty
	if err := m.Unmarshal("val", &b); err != nil || b == nil || b.Id != 2 || b.Val != 200 {
		t.Fatalf("value -> pointer: got %+v, err %v", b, err)
	}
	// 同型读取：原行为
	var c *TestEquipProperty
	if err := m.Unmarshal("ptr", &c); err != nil || c == nil || c.Id != 1 {
		t.Fatalf("pointer -> pointer: got %+v, err %v", c, err)
	}
	var d TestEquipProperty
	if err := m.Unmarshal("val", &d); err != nil || d.Id != 2 {
		t.Fatalf("value -> value: got %+v, err %v", d, err)
	}

	// 读入目标是接口：原行为（指针原样装进接口，不被解引用）
	var e any
	if err := m.Unmarshal("ptr", &e); err != nil {
		t.Fatalf("pointer -> interface: %v", err)
	}
	p, ok := e.(*TestEquipProperty)
	if !ok || p.Id != 1 {
		t.Fatalf("pointer -> interface: got %#v", e)
	}

	// 深拷贝：改读出的结果不影响所存的值
	a.Id, b.Id, c.Val, p.Val = 9, 9, 9, 9
	var again *TestEquipProperty
	var again2 TestEquipProperty
	if err := m.Unmarshal("ptr", &again); err != nil || again.Id != 1 || again.Val != 100 {
		t.Fatalf("deep copy broken (ptr): got %+v, err %v", again, err)
	}
	if err := m.Unmarshal("val", &again2); err != nil || again2.Id != 2 {
		t.Fatalf("deep copy broken (val): got %+v, err %v", again2, err)
	}

	// 只读兼容，不改变所存形态
	if _, ok := m["ptr"].(*TestEquipProperty); !ok {
		t.Fatalf("stored form changed: %T", m["ptr"])
	}
	if _, ok := m["val"].(TestEquipProperty); !ok {
		t.Fatalf("stored form changed: %T", m["val"])
	}

	// 存的是空指针 = 没有值：目标保持不变
	m.Set("nil", (*TestEquipProperty)(nil))
	keep := TestEquipProperty{Id: 7}
	if err := m.Unmarshal("nil", &keep); err != nil || keep.Id != 7 {
		t.Fatalf("nil pointer: got %+v, err %v", keep, err)
	}

	// 类型确实不同：仍然报错
	if err := m.Unmarshal("ptr", &other); err == nil {
		t.Fatal("pointer -> other struct: expected error")
	}
	if err := m.Unmarshal("val", &other); err == nil {
		t.Fatal("value -> other struct: expected error")
	}
	if err := m.Unmarshal("ptr", &[]TestEquipProperty{}); err == nil {
		t.Fatal("pointer -> slice: expected error")
	}

	// 只对齐一层指针，且对齐之后必须真的能赋值：对齐不了的报错，不能走到 reflect.Set 去 panic
	t.Run("align must be exact", func(t *testing.T) {
		for _, read := range []struct {
			name string
			f    func(k string, i any) error
		}{{"Unmarshal", m.Unmarshal}, {"Borrow", m.Borrow}} {
			pp := &TestEquipProperty{Id: 1, Val: 100}
			m.Set("pp", &pp) //**TestEquipProperty
			if err := read.f("pp", &pp); err == nil {
				t.Fatalf("%s: **T -> *T 应只对齐一层并报错", read.name)
			}
			// 匿名结构体与具名结构体底层相同，值之间可赋值，但它们的指针之间不可赋值
			type named struct {
				Id  int32
				Val int64
			}
			m.Set("anon", struct {
				Id  int32
				Val int64
			}{Id: 2, Val: 200})
			var np *named
			if err := read.f("anon", &np); err == nil {
				t.Fatalf("%s: 匿名结构体 -> *具名结构体 应报错", read.name)
			}
			var nv named
			if err := read.f("anon", &nv); err != nil || nv.Id != 2 {
				t.Fatalf("%s: 匿名结构体 -> 具名结构体（同型可赋值）应成功: %+v, err %v", read.name, nv, err)
			}
		}
	})
}

// TestAttachBorrowReadOnly Borrow 与 Unmarshal 同形但不深拷贝：切片 / map / 指针与缓存共享，所以只读；
// Unmarshal 仍是深拷贝。两者都兼容指针 / 值两种存法。
func TestAttachBorrowReadOnly(t *testing.T) {
	type state struct {
		Name  string
		Items []int32
		Meta  map[string]int32
		Inner *TestEquipProperty
	}
	newState := func() state {
		return state{Name: "a", Items: []int32{1, 2}, Meta: map[string]int32{"x": 1}, Inner: &TestEquipProperty{Id: 1}}
	}
	forms := map[string]func() any{
		"value":   func() any { return newState() },
		"pointer": func() any { s := newState(); return &s },
	}
	for name, build := range forms {
		t.Run(name, func(t *testing.T) {
			m := make(Values)
			m.Set("k", build())

			var b state
			if err := m.Borrow("k", &b); err != nil {
				t.Fatalf("Borrow: %v", err)
			}
			if b.Name != "a" || len(b.Items) != 2 || b.Meta["x"] != 1 || b.Inner == nil || b.Inner.Id != 1 {
				t.Fatalf("Borrow result: %+v", b)
			}
			if n := testing.AllocsPerRun(100, func() { _ = m.Borrow("k", &b) }); n != 0 {
				t.Fatalf("Borrow 热路径应零分配，实得 %v", n)
			}

			// 共享：改 Borrow 读出的切片元素 / map / 指向的对象，缓存跟着变
			b.Items[0], b.Meta["x"], b.Inner.Val = 9, 9, 9
			var c state
			if err := m.Borrow("k", &c); err != nil {
				t.Fatalf("Borrow again: %v", err)
			}
			if c.Items[0] != 9 || c.Meta["x"] != 9 || c.Inner.Val != 9 {
				t.Fatalf("Borrow 应与缓存共享内部引用: %+v", c)
			}
			// 结构体本身是各自的拷贝：标量字段互不影响
			b.Name = "changed"
			if c.Name != "a" {
				t.Fatalf("Borrow 的结构体应是浅拷贝: %q", c.Name)
			}

			// Unmarshal 仍是深拷贝：改读出的值不影响缓存
			var d state
			if err := m.Unmarshal("k", &d); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			d.Items[1], d.Inner.Id = 7, 7
			d.Meta["y"] = 7
			var e state
			if err := m.Unmarshal("k", &e); err != nil {
				t.Fatalf("Unmarshal again: %v", err)
			}
			if e.Items[1] != 2 || e.Inner.Id != 1 || len(e.Meta) != 1 {
				t.Fatalf("Unmarshal 应是深拷贝: %+v", e)
			}
		})
	}

	// 缓存是值、读入指针：给出一份浅拷贝的地址；缓存是指针、读入指针：同一个指针
	t.Run("pointer target", func(t *testing.T) {
		m := make(Values)
		m.Set("v", TestEquipProperty{Id: 1, Val: 100})
		var p *TestEquipProperty
		if err := m.Borrow("v", &p); err != nil || p == nil || p.Id != 1 {
			t.Fatalf("value -> pointer: %+v, err %v", p, err)
		}
		q := &TestEquipProperty{Id: 2, Val: 200}
		m.Set("p", q)
		var r *TestEquipProperty
		if err := m.Borrow("p", &r); err != nil || r != q {
			t.Fatalf("pointer -> pointer 应是同一个指针: %v", err)
		}
	})

	// 冷路径：解码后缓存，之后零分配
	t.Run("cold path", func(t *testing.T) {
		m := make(Values)
		m["k"] = bson.D{{Key: "id", Value: int32(1)}, {Key: "val", Value: int64(100)}}
		var got TestEquipProperty
		if err := m.Borrow("k", &got); err != nil || got.Id != 1 || got.Val != 100 {
			t.Fatalf("cold Borrow: %+v, err %v", got, err)
		}
		if n := testing.AllocsPerRun(100, func() { _ = m.Borrow("k", &got) }); n != 0 {
			t.Fatalf("冷路径之后的 Borrow 应零分配，实得 %v", n)
		}
	})

	// 缺键 / 空指针：目标不变、不报错；类型确实不同：报错
	t.Run("absent and mismatch", func(t *testing.T) {
		m := make(Values)
		m.Set("nil", (*TestEquipProperty)(nil))
		m.Set("p", &TestEquipProperty{Id: 1})
		keep := TestEquipProperty{Id: 7}
		if err := m.Borrow("absent", &keep); err != nil || keep.Id != 7 {
			t.Fatalf("absent: %+v, err %v", keep, err)
		}
		if err := m.Borrow("nil", &keep); err != nil || keep.Id != 7 {
			t.Fatalf("nil pointer: %+v, err %v", keep, err)
		}
		if err := m.Borrow("p", &struct{ X int }{}); err == nil {
			t.Fatal("type mismatch: expected error")
		}
	})
}

// TestAttachLoadColdPathCache 冷路径读完后把解码结果记入缓存：Unmarshal 记入与调用方目标独立的副本，
// Borrow 直接记入那次读出的对象（不拷贝，共享）。Load 是两者共用的实现，deep 决定是否拷贝。
func TestAttachLoadColdPathCache(t *testing.T) {
	raw := func() Values {
		return Values{"k": bson.D{{Key: "id", Value: int32(1)}, {Key: "val", Value: int64(100)}}}
	}

	// Unmarshal：缓存与读出的值互不影响
	m := raw()
	var a *TestEquipProperty
	if err := m.Unmarshal("k", &a); err != nil || a == nil || a.Id != 1 {
		t.Fatalf("cold Unmarshal: %v, err %v", a, err)
	}
	if cached, ok := m["k"].(*TestEquipProperty); !ok || cached == a {
		t.Fatalf("Unmarshal 冷路径应记入独立的副本: %#v", m["k"])
	}
	a.Id = 9
	var b *TestEquipProperty
	if err := m.Unmarshal("k", &b); err != nil || b.Id != 1 {
		t.Fatalf("缓存不应被读出的值影响: %+v, err %v", b, err)
	}

	// Borrow：直接记入那次读出的对象
	m = raw()
	var c *TestEquipProperty
	if err := m.Borrow("k", &c); err != nil || c == nil || c.Id != 1 {
		t.Fatalf("cold Borrow: %v, err %v", c, err)
	}
	if m["k"] != any(c) {
		t.Fatalf("Borrow 冷路径应直接记入读出的对象: %#v", m["k"])
	}

	// Load：deep 与两个方法等价
	var d, e TestEquipProperty
	if err := raw().Load("k", &d, true); err != nil || d.Id != 1 || d.Val != 100 {
		t.Fatalf("Load deep: %+v, err %v", d, err)
	}
	if err := raw().Load("k", &e, false); err != nil || e.Id != 1 || e.Val != 100 {
		t.Fatalf("Load shallow: %+v, err %v", e, err)
	}

	// 旧版 JSON 字节的冷路径同样记入缓存，之后零分配
	j := Values{"k": []byte(`{"id":2,"val":200}`)}
	var g TestEquipProperty
	if err := j.Borrow("k", &g); err != nil || g.Id != 2 || g.Val != 200 {
		t.Fatalf("cold Borrow from JSON: %+v, err %v", g, err)
	}
	if n := testing.AllocsPerRun(100, func() { _ = j.Borrow("k", &g) }); n != 0 {
		t.Fatalf("JSON 冷路径之后的 Borrow 应零分配，实得 %v", n)
	}
}

// BenchmarkAttachRead 同一份缓存值的两种读法：Unmarshal 深拷贝，Borrow 不拷贝。
func BenchmarkAttachRead(b *testing.B) {
	type state struct {
		Completed uint32
		Claimed   []int32
	}
	props := []*TestEquipProperty{{Id: 1, Val: 100}, {Id: 2, Val: 200}, {Id: 3, Val: 300}, {Id: 4, Val: 400}}
	cases := []struct {
		name string
		v    any
		out  func() any
	}{
		{"struct", state{Completed: 0xfffe, Claimed: []int32{1, 2, 3, 4, 5, 6, 7, 8, 9}}, func() any { return &state{} }},
		{"slice_of_pointers", props, func() any { return &[]*TestEquipProperty{} }},
	}
	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			m := make(Values)
			m.Set("k", c.v)
			reads := []struct {
				name string
				f    func(k string, i any) error
			}{{"Unmarshal", m.Unmarshal}, {"Borrow", m.Borrow}}
			for _, read := range reads {
				b.Run(read.name, func(b *testing.B) {
					out := c.out()
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						if err := read.f("k", out); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		})
	}
}
