package schema

import (
	"testing"
)

type unsetLeaf struct {
	Lv   int32
	Name string
}

type unsetRoot struct {
	BreakLv    int32
	Goods      map[int32]int64
	SoulRelics map[int32]*unsetLeaf
	Tags       map[string]string
	Slots      []*unsetLeaf
	Profile    *unsetLeaf
	ByValue    map[int32]unsetLeaf //值类型 map：Go 本身就不允许就地改元素
}

func newUnsetRoot() *unsetRoot {
	return &unsetRoot{
		BreakLv:    5,
		Goods:      map[int32]int64{10001: 99, 10002: 7},
		SoulRelics: map[int32]*unsetLeaf{1: {Lv: 7, Name: "a"}, 2: {Lv: 8}},
		Tags:       map[string]string{"AbC": "x", "d": "y"},
		Slots:      []*unsetLeaf{{Lv: 1}, {Lv: 2}},
		Profile:    &unsetLeaf{Lv: 3, Name: "p"},
		ByValue:    map[int32]unsetLeaf{1: {Lv: 9}},
	}
}

func mustUnset(t *testing.T, sch *Schema, m *unsetRoot, path string) {
	t.Helper()
	if err := sch.Unset(m, path); err != nil {
		t.Fatalf("Unset(%q) 不该报错: %v", path, err)
	}
}

// 整字段置零 —— 原本就能工作，回归守卫
func TestUnsetWholeField(t *testing.T) {
	sch, _ := Parse(&unsetRoot{})
	m := newUnsetRoot()

	mustUnset(t, sch, m, "BreakLv")
	if m.BreakLv != 0 {
		t.Fatalf("BreakLv 应置零，实际 %d", m.BreakLv)
	}
	mustUnset(t, sch, m, "Profile")
	if m.Profile != nil {
		t.Fatalf("Profile 应置 nil，实际 %+v", m.Profile)
	}
}

// 🔴 map 子键必须真的从 map 里删掉。
//
// 改动前走 SetValue，LookUpField("goods.10001") 必然为 nil、报 "field not exist"，
// 内存纹丝不动；而落库的 $unset 照常执行 —— 内存与库分叉，常驻内存的模型一直错到重启。
func TestUnsetMapKey(t *testing.T) {
	sch, _ := Parse(&unsetRoot{})
	m := newUnsetRoot()

	mustUnset(t, sch, m, "Goods.10001")
	if _, ok := m.Goods[10001]; ok {
		t.Fatalf("Goods.10001 应被删除，实际 %v", m.Goods)
	}
	if m.Goods[10002] != 7 {
		t.Fatalf("同 map 内其它键不该受影响，实际 %v", m.Goods)
	}

	mustUnset(t, sch, m, "SoulRelics.1")
	if _, ok := m.SoulRelics[1]; ok {
		t.Fatalf("SoulRelics.1 应被删除，实际 %v", m.SoulRelics)
	}
	if m.SoulRelics[2] == nil {
		t.Fatal("SoulRelics.2 不该受影响")
	}

	//字符串键，大小写照原样匹配
	mustUnset(t, sch, m, "Tags.AbC")
	if _, ok := m.Tags["AbC"]; ok {
		t.Fatalf("Tags.AbC 应被删除，实际 %v", m.Tags)
	}
	if m.Tags["d"] != "y" {
		t.Fatal("Tags.d 不该受影响")
	}
}

// 穿过 map 键继续下钻到值类型的字段，置零而不是删键
func TestUnsetDeepThroughMap(t *testing.T) {
	sch, _ := Parse(&unsetRoot{})
	m := newUnsetRoot()

	mustUnset(t, sch, m, "SoulRelics.1.Lv")
	if m.SoulRelics[1] == nil {
		t.Fatal("只清字段，不该把整个 map 元素删掉")
	}
	if m.SoulRelics[1].Lv != 0 {
		t.Fatalf("SoulRelics.1.Lv 应置零，实际 %d", m.SoulRelics[1].Lv)
	}
	if m.SoulRelics[1].Name != "a" {
		t.Fatalf("同元素内其它字段不该受影响，实际 %q", m.SoulRelics[1].Name)
	}
}

// 嵌套结构体字段
func TestUnsetNestedStruct(t *testing.T) {
	sch, _ := Parse(&unsetRoot{})
	m := newUnsetRoot()

	mustUnset(t, sch, m, "Profile.Name")
	if m.Profile.Name != "" {
		t.Fatalf("Profile.Name 应置零，实际 %q", m.Profile.Name)
	}
	if m.Profile.Lv != 3 {
		t.Fatal("Profile.Lv 不该受影响")
	}
}

// slice 下标：置零值，长度不变（与 mongo $unset 对数组的语义一致）
func TestUnsetSliceIndex(t *testing.T) {
	sch, _ := Parse(&unsetRoot{})
	m := newUnsetRoot()

	mustUnset(t, sch, m, "Slots.0")
	if len(m.Slots) != 2 {
		t.Fatalf("长度不该变，实际 %d", len(m.Slots))
	}
	if m.Slots[0] != nil {
		t.Fatalf("Slots.0 应置零，实际 %+v", m.Slots[0])
	}
	if m.Slots[1] == nil || m.Slots[1].Lv != 2 {
		t.Fatal("Slots.1 不该受影响")
	}

	mustUnset(t, sch, m, "Slots.1.Lv")
	if m.Slots[1].Lv != 0 {
		t.Fatalf("Slots.1.Lv 应置零，实际 %d", m.Slots[1].Lv)
	}
}

// 路径中断（nil 指针 / nil map / 不存在的键 / 越界）视为「本来就没有」，不报错
func TestUnsetMissingIsNoop(t *testing.T) {
	sch, _ := Parse(&unsetRoot{})

	m := &unsetRoot{} //全 nil
	for _, p := range []string{"Goods.1", "SoulRelics.1.Lv", "Profile.Lv", "Slots.0", "Slots.5.Lv"} {
		if err := sch.Unset(m, p); err != nil {
			t.Errorf("空对象上 Unset(%q) 不该报错: %v", p, err)
		}
	}

	full := newUnsetRoot()
	for _, p := range []string{"Goods.99999", "SoulRelics.99.Lv", "Slots.9"} {
		if err := sch.Unset(full, p); err != nil {
			t.Errorf("不存在的键 Unset(%q) 不该报错: %v", p, err)
		}
	}
	if len(full.Goods) != 2 || len(full.SoulRelics) != 2 {
		t.Fatal("不存在的键不该动到已有数据")
	}
}

// 路径写错必须报错，而不是静默无事发生
func TestUnsetBadPath(t *testing.T) {
	sch, _ := Parse(&unsetRoot{})
	m := newUnsetRoot()

	bad := map[string]string{
		"":                  "空路径",
		"NoSuchField":       "字段不存在",
		"NoSuchField.1":     "根字段不存在",
		"Profile.NoSuch":    "嵌套结构体里没有这个字段",
		"SoulRelics.1.Nope": "map 值结构体里没有这个字段",
		"BreakLv.1":         "标量不能再往下钻",
		"Goods.abc":         "map 键类型对不上",
		"Goods.":            "空段",
	}
	for p, why := range bad {
		if err := sch.Unset(m, p); err == nil {
			t.Errorf("Unset(%q) 应报错(%s)", p, why)
		}
	}
}

// map[K]Struct（值类型）的元素在 Go 里本就不可就地修改，
// reflect 会 panic —— 必须被兜成错误，不能把调用方带崩
func TestUnsetUnaddressableRecovered(t *testing.T) {
	sch, _ := Parse(&unsetRoot{})
	m := newUnsetRoot()

	err := sch.Unset(m, "ByValue.1.Lv")
	if err == nil {
		t.Fatal("值类型 map 元素的字段不可就地清除，应报错")
	}
	//删整个键是可以的
	if err = sch.Unset(m, "ByValue.1"); err != nil {
		t.Fatalf("删值类型 map 的整个键应当可行: %v", err)
	}
	if _, ok := m.ByValue[1]; ok {
		t.Fatal("ByValue.1 应被删除")
	}
}
