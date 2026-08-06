package schema

import "testing"

func newNameBenchSchema(tb testing.TB) *Schema {
	s, err := GetOrParse(&dbnameRoot{}, nil)
	if err != nil {
		tb.Fatal(err)
	}
	return s
}

// BenchmarkName_Baseline 改动前单段路径的等价开销：一次 map 查找 + 取缓存名。
func BenchmarkName_Baseline(b *testing.B) {
	s := newNameBenchSchema(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if f := s.LookUpField("BreakLv"); f != nil {
			_ = f.DBName()
		}
	}
}

// BenchmarkName_Single 单段路径走 GetName 的快路径（比 baseline 多一次 strings.Contains）。
func BenchmarkName_Single(b *testing.B) {
	s := newNameBenchSchema(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.DBName("BreakLv")
	}
}

// BenchmarkName_MapKey 两段：字段 + map 键，业务上最常见的子键写法（goods.10001）。
func BenchmarkName_MapKey(b *testing.B) {
	s := newNameBenchSchema(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.DBName("SoulRelics.1")
	}
}

// BenchmarkName_Nested 两段：字段 + 嵌套结构体字段（bless.lv）。
func BenchmarkName_Nested(b *testing.B) {
	s := newNameBenchSchema(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.DBName("Profile.Lv")
	}
}

// BenchmarkName_Deep 三段：字段 + map 键 + 值类型字段，需要按类型解析子 schema。
func BenchmarkName_Deep(b *testing.B) {
	s := newNameBenchSchema(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.DBName("SoulRelics.1.Lv")
	}
}

// 下面三个是**业务实际写法**：调用方本来就按落库名写（goods.10001 / bless.lv），
// 整条路径逐字不变，Builder 不启用，直接返回入参 —— 应当零分配。
// 上面几个用 PascalCase 入参，走的是「需要改名」分支，是最坏情况。

func BenchmarkName_MapKey_Canonical(b *testing.B) {
	s := newNameBenchSchema(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.DBName("soulrelics.1")
	}
}

func BenchmarkName_Nested_Canonical(b *testing.B) {
	s := newNameBenchSchema(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.DBName("profile.lv")
	}
}

func BenchmarkName_Deep_Canonical(b *testing.B) {
	s := newNameBenchSchema(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.DBName("soulrelics.1.lv")
	}
}
