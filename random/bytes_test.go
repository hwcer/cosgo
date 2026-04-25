package random

import (
	"strings"
	"testing"
)

func TestBytes_NewLength(t *testing.T) {
	for _, l := range []int{1, 8, 16, 64} {
		got := Strings.New(l)
		if len(got) != l {
			t.Errorf("want len %d, got %d", l, len(got))
		}
	}
}

func TestBytes_UsesCharset(t *testing.T) {
	charset := []byte("0123456789abcdefghijklmnopqrstuvwxyz")
	b := NewBytes(charset)
	got := b.New(256)
	for _, c := range got {
		if !strings.ContainsRune(string(charset), rune(c)) {
			t.Errorf("byte %q not in charset", c)
		}
	}
}

// TestBytes_NotPredictable 粗略检测:两个 32 字节样本不应相同。
// 修复前 math/rand + time seed,大量调用仍可能碰撞;新版 crypto/rand 不应该。
func TestBytes_NotPredictable(t *testing.T) {
	a := Strings.String(32)
	c := Strings.String(32)
	if a == c {
		t.Errorf("two samples unexpectedly equal: %s vs %s", a, c)
	}
}

func TestBytes_ZeroLength(t *testing.T) {
	if got := Strings.New(0); got != nil {
		t.Errorf("New(0) should return nil, got %v", got)
	}
}

func TestBytes_SingleCharCharset(t *testing.T) {
	b := NewBytes([]byte("X"))
	got := b.String(8)
	if got != "XXXXXXXX" {
		t.Errorf("single-char charset should fill, got %q", got)
	}
}

func TestRelativeMulti_DoesNotMutateInput(t *testing.T) {
	items := map[int32]int32{1: 100, 2: 200, 3: 300}
	original := make(map[int32]int32, len(items))
	for k, v := range items {
		original[k] = v
	}
	_ = RelativeMulti(items, 2)
	if len(items) != len(original) {
		t.Errorf("input map mutated: len=%d, want %d", len(items), len(original))
	}
	for k, v := range original {
		if items[k] != v {
			t.Errorf("input map mutated: key=%d, got=%d, want=%d", k, items[k], v)
		}
	}
}

// ==================== Benchmark ====================

func BenchmarkBytes_String6(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Strings.String(6)
	}
}

func BenchmarkBytes_String32(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Strings.String(32)
	}
}

func BenchmarkRoll(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Roll(1, 10000)
	}
}

func BenchmarkRelative(b *testing.B) {
	items := map[int32]int32{1: 100, 2: 200, 3: 300, 4: 400}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Relative(items)
	}
}
