package schema

import (
	"errors"
	"testing"
	"time"
)

// 具名自引用：树/链表节点最常见的形状
type CycNode struct {
	ID    int64 `bson:"_id"`
	Name  string
	Child *CycNode
}

// 互引用：环上没有任何一个类型「是它自己」，只判 self 的实现会漏掉
type CycOrder struct {
	ID   int64 `bson:"_id"`
	User *CycUser
}

type CycUser struct {
	ID    int64 `bson:"_id"`
	Order *CycOrder
}

// 三角环，验证链上任意深度都能识别
type CycTriA struct {
	ID int64 `bson:"_id"`
	B  *CycTriB
}
type CycTriB struct{ C *CycTriC }
type CycTriC struct{ A *CycTriA }

// map 元素自引用：元素类型不参与 eager 解析，本就正常，作为回归守卫
type CycMap struct {
	ID       int64 `bson:"_id"`
	Children map[int32]*CycMap
}

// 兄弟字段：同一层有多个结构体字段，其中一个成环、另一个是普通子结构。
// 用来压解析链的复制语义 —— 若兄弟递归共享底层数组，后者会覆写前者的链。
type CycSibRoot struct {
	ID    int64 `bson:"_id"`
	Deep  *CycSibL1
	Plain *CycSibPlain
}
type CycSibL1 struct {
	L2    *CycSibL2
	Plain *CycSibPlain
}
type CycSibL2 struct {
	Back  *CycSibRoot // 回到最外层
	Plain *CycSibPlain
}
type CycSibPlain struct {
	Val int64 `bson:"val"`
}

// 匿名自嵌入：字段要被提升到父结构上，成环即无限提升，必须报错
type CycEmbedSelf struct {
	ID int64 `bson:"_id"`
	*CycEmbedSelf
}

// 匿名互嵌入
type CycEmbedA struct {
	ID int64 `bson:"_id"`
	*CycEmbedB
}
type CycEmbedB struct {
	*CycEmbedA
}

// withShortTimeout 把 SchemaInitTimeout 压到很短：一旦环检测退化回「等 chan」，
// 测试会立刻失败而不是挂满 30 秒。
func withShortTimeout(t *testing.T) {
	t.Helper()
	old := SchemaInitTimeout
	SchemaInitTimeout = 300 * time.Millisecond
	t.Cleanup(func() { SchemaInitTimeout = old })
}

// mustParseFast 解析必须成功且几乎瞬时（撞上 waitSchemaInit 就会超过阈值）。
func mustParseFast(t *testing.T, dest any) *Schema {
	t.Helper()
	t0 := time.Now()
	s, err := GetOrParse(dest, New())
	cost := time.Since(t0)
	if err != nil {
		t.Fatalf("%T 解析失败: %v", dest, err)
	}
	if s == nil {
		t.Fatalf("%T 解析返回 nil schema", dest)
	}
	if cost > 100*time.Millisecond {
		t.Fatalf("%T 解析耗时 %v，说明又走进了 waitSchemaInit 等待", dest, cost)
	}
	return s
}

// 具名自引用必须正常解析，且 Embedded 指回自己那张 Schema。
//
// 改动前：ParseField 无脑 GetOrParse，撞上缓存里属于自己的 in-progress 记录，
// 然后在 waitSchemaInit 里等一个只有自己能 close 的 chan —— 满 30 秒才失败。
func TestParseCycleNamedSelf(t *testing.T) {
	withShortTimeout(t)
	s := mustParseFast(t, &CycNode{})

	f := s.LookUpField("Child")
	if f == nil {
		t.Fatal("Child 字段丢失")
	}
	if f.Embedded != s {
		t.Fatalf("Child.Embedded 应指回自身 Schema，实际 %v", f.Embedded)
	}
	//自引用不该影响正常字段的解析结果
	if got, err := s.DBName("Name"); err != nil || got != "name" {
		t.Fatalf("DBName(Name) = %q,%v，期望 name", got, err)
	}
	//多级路径能穿过自引用继续走
	if got, err := s.DBName("Child.Child.Name"); err != nil || got != "child.child.name" {
		t.Fatalf("DBName(Child.Child.Name) = %q,%v，期望 child.child.name", got, err)
	}
}

// 互引用：环上没有「自己引用自己」，只判 self 的实现会漏。
func TestParseCycleMutual(t *testing.T) {
	withShortTimeout(t)
	s := mustParseFast(t, &CycOrder{})

	if got, err := s.DBName("User.Order.User.ID"); err != nil || got != "user.order.user._id" {
		t.Fatalf("DBName 穿环失败: %q,%v", got, err)
	}
}

// 三角环 A→B→C→A：链上任意深度都要能识别。
func TestParseCycleTriangle(t *testing.T) {
	withShortTimeout(t)
	s := mustParseFast(t, &CycTriA{})

	if got, err := s.DBName("B.C.A.ID"); err != nil || got != "b.c.a._id" {
		t.Fatalf("DBName 穿三角环失败: %q,%v", got, err)
	}
}

// map 元素自引用本来就正常（元素类型不 eager 解析），守住别被改坏。
func TestParseCycleMapElement(t *testing.T) {
	withShortTimeout(t)
	s := mustParseFast(t, &CycMap{})

	if got, err := s.DBName("Children.7.ID"); err != nil || got != "children.7._id" {
		t.Fatalf("DBName 穿 map 元素自引用失败: %q,%v", got, err)
	}
}

// 同一层多个结构体字段：一个成深环、一个是普通子结构。
// 解析链若在兄弟递归之间共享底层数组，第二个兄弟会覆写第一个的链，环检测就会漏判或误判。
func TestParseCycleSiblingFields(t *testing.T) {
	withShortTimeout(t)
	s := mustParseFast(t, &CycSibRoot{})

	//深环能穿回最外层
	if got, err := s.DBName("Deep.L2.Back.ID"); err != nil || got != "deep.l2.back._id" {
		t.Fatalf("穿深环失败: %q,%v", got, err)
	}
	//非环的兄弟字段在各层都要解析完整
	for path, want := range map[string]string{
		"Plain.Val":          "plain.val",
		"Deep.Plain.Val":     "deep.plain.val",
		"Deep.L2.Plain.Val":  "deep.l2.plain.val",
		"Deep.L2.Back.Plain": "deep.l2.back.plain",
	} {
		if got, err := s.DBName(path); err != nil || got != want {
			t.Errorf("DBName(%q) = %q,%v，期望 %q", path, got, err, want)
		}
	}
}

// 匿名嵌入成环必须**报错**，而不是超时、也不是返回字段残缺的半成品。
func TestParseCycleEmbeddedRejected(t *testing.T) {
	withShortTimeout(t)

	for _, dest := range []any{&CycEmbedSelf{}, &CycEmbedA{}} {
		t0 := time.Now()
		s, err := GetOrParse(dest, New())
		cost := time.Since(t0)
		if err == nil {
			t.Errorf("%T 匿名嵌入成环应报错，实际成功: %v", dest, s)
			continue
		}
		if !errors.Is(err, ErrEmbeddedCycle) {
			t.Errorf("%T 期望 ErrEmbeddedCycle，实际 %v", dest, err)
		}
		if cost > 100*time.Millisecond {
			t.Errorf("%T 应立即报错，实际耗时 %v（说明走进了等待）", dest, cost)
		}
	}
}
