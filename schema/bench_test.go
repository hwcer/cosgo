package schema

import (
	"testing"
)

type benchUser struct {
	ID      int64   `json:"id" bson:"_id"`
	Name    string  `json:"name" bson:"name"`
	Age     int     `json:"age" bson:"age"`
	Email   string  `json:"email" bson:"email"`
	Balance float64 `json:"balance" bson:"balance"`
}

func newBenchSchema(tb testing.TB) *Schema {
	s, err := GetOrParse(&benchUser{}, nil)
	if err != nil {
		tb.Fatal(err)
	}
	return s
}

// BenchmarkLookUpField_GoName 按 Go 字段名查找(第一优先级,原实现也是第一次 map 命中)。
func BenchmarkLookUpField_GoName(b *testing.B) {
	s := newBenchSchema(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.LookUpField("Name")
	}
}

// BenchmarkLookUpField_DBName 按 db 标签查找(原实现需要 2 次 map 查询,现在 1 次)。
func BenchmarkLookUpField_DBName(b *testing.B) {
	s := newBenchSchema(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.LookUpField("_id")
	}
}

// BenchmarkLookUpField_Miss 三次 map 全部 miss(原实现 3 次查询,现在 1 次)。
func BenchmarkLookUpField_Miss(b *testing.B) {
	s := newBenchSchema(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.LookUpField("nonexistent")
	}
}

// BenchmarkGetValue_Single 单级字段取值(应命中 common case 快路径,不分配合并 slice)。
func BenchmarkGetValue_Single(b *testing.B) {
	s := newBenchSchema(b)
	u := &benchUser{Name: "alice", Age: 30}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.GetValue(u, "Name")
	}
}

// BenchmarkSetValue_Single 单级字段赋值(应命中 common case 快路径,不分配合并 slice)。
func BenchmarkSetValue_Single(b *testing.B) {
	s := newBenchSchema(b)
	u := &benchUser{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.SetValue(u, "alice", "Name")
	}
}

// BenchmarkParse_CacheHit 默认 Parse 缓存命中（reflect.Type 做 key，零 boxing）
// 预分配 user 避免循环内 &benchUser{} 逃逸到堆的干扰
func BenchmarkParse_CacheHit(b *testing.B) {
	u := &benchUser{}
	_, err := Parse(u)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(u)
	}
}

// BenchmarkParseWithSpecialTableName_CacheHit 带自定义表名的缓存命中（struct key 需 boxing）
func BenchmarkParseWithSpecialTableName_CacheHit(b *testing.B) {
	opts := New()
	_, err := opts.ParseWithSpecialTableName(&benchUser{}, "custom")
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = opts.ParseWithSpecialTableName(&benchUser{}, "custom")
	}
}
