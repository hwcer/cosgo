package safety

import (
	"fmt"
	"sync"
	"testing"
)

func TestParseIPv4(t *testing.T) {
	tests := []struct {
		input string
		want  uint32
	}{
		{"0.0.0.0", 0},
		{"0.0.0.1", 1},
		{"127.0.0.1", 0x7F000001},
		{"192.168.1.2", 0xC0A80102},
		{"10.0.0.0", 0x0A000000},
		{"255.255.255.255", 0xFFFFFFFF},
		{"192.168.1.2:8080", 0xC0A80102}, // 端口被剥离
		{"", 0},
		{"abc", 0},
		{"256.0.0.0", 0},
		{"1.2.3", 0},
	}
	for _, tt := range tests {
		got := parseIPv4(tt.input)
		if got != tt.want {
			t.Errorf("parseIPv4(%q) = 0x%08X, want 0x%08X", tt.input, got, tt.want)
		}
	}
}

func TestParseRule(t *testing.T) {
	tests := []struct {
		rule       string
		wantStart  uint32
		wantEnd    uint32
	}{
		// 单 IP
		{"192.168.1.1", 0xC0A80101, 0xC0A80101},
		// 范围
		{"10.0.0.0~10.255.255.255", 0x0A000000, 0x0AFFFFFF},
		// CIDR
		{"10.0.0.0/8", 0x0A000000, 0x0AFFFFFF},
		{"172.16.0.0/12", 0xAC100000, 0xAC1FFFFF},
		{"192.168.0.0/16", 0xC0A80000, 0xC0A8FFFF},
		{"192.168.1.0/24", 0xC0A80100, 0xC0A801FF},
		{"192.168.1.128/32", 0xC0A80180, 0xC0A80180},
	}
	for _, tt := range tests {
		start, end := parseRule(tt.rule)
		if start != tt.wantStart || end != tt.wantEnd {
			t.Errorf("parseRule(%q) = (0x%08X, 0x%08X), want (0x%08X, 0x%08X)",
				tt.rule, start, end, tt.wantStart, tt.wantEnd)
		}
	}
}

func TestMatchSingleIP(t *testing.T) {
	s := New()
	s.Update("test", "192.168.1.100", StatusDisable, false)

	if got := s.Match("192.168.1.100", false); got != StatusDisable {
		t.Errorf("exact match: got %d, want %d", got, StatusDisable)
	}
	if got := s.Match("192.168.1.101", false); got != StatusNone {
		t.Errorf("miss: got %d, want %d", got, StatusNone)
	}
}

func TestMatchRange(t *testing.T) {
	s := New()
	s.Update("office", "10.0.0.0~10.0.0.255", StatusEnable, false)

	if got := s.Match("10.0.0.1", false); got != StatusEnable {
		t.Errorf("in range: got %d, want %d", got, StatusEnable)
	}
	if got := s.Match("10.0.0.255", false); got != StatusEnable {
		t.Errorf("range end: got %d, want %d", got, StatusEnable)
	}
	if got := s.Match("10.0.1.0", false); got != StatusNone {
		t.Errorf("out of range: got %d, want %d", got, StatusNone)
	}
}

func TestMatchCIDR(t *testing.T) {
	s := New()
	s.Update("subnet", "172.16.0.0/12", StatusEnable, false)

	if got := s.Match("172.16.0.1", false); got != StatusEnable {
		t.Errorf("in CIDR: got %d, want %d", got, StatusEnable)
	}
	if got := s.Match("172.31.255.254", false); got != StatusEnable {
		t.Errorf("CIDR end: got %d, want %d", got, StatusEnable)
	}
	if got := s.Match("172.32.0.0", false); got != StatusNone {
		t.Errorf("out of CIDR: got %d, want %d", got, StatusNone)
	}
}

func TestUseLocalAddress(t *testing.T) {
	s := New()
	s.UseLocalAddress()

	locals := []string{"127.0.0.1", "10.1.2.3", "172.16.0.1", "192.168.0.1"}
	for _, ip := range locals {
		if got := s.Match(ip, true); got != StatusEnable {
			t.Errorf("local %s: got %d, want %d", ip, got, StatusEnable)
		}
	}
	// local 规则在 useLocalAddress=false 时不生效
	if got := s.Match("10.1.2.3", false); got != StatusNone {
		t.Errorf("local disabled: got %d, want %d", got, StatusNone)
	}
}

func TestDeleteRule(t *testing.T) {
	s := New()
	s.Update("block", "1.2.3.4", StatusDisable, false)

	if got := s.Match("1.2.3.4", false); got != StatusDisable {
		t.Fatal("should match before delete")
	}

	s.Delete("block")

	if got := s.Match("1.2.3.4", false); got != StatusNone {
		t.Errorf("should not match after delete: got %d", got)
	}
}

func TestDeleteNonExistent(t *testing.T) {
	s := New()
	s.Update("a", "1.1.1.1", StatusEnable, false)
	s.Delete("nonexistent") // 不应 panic 或影响已有规则

	if got := s.Match("1.1.1.1", false); got != StatusEnable {
		t.Error("existing rule should survive delete of nonexistent")
	}
}

func TestReload(t *testing.T) {
	s := New()
	s.Update("old", "1.1.1.1", StatusDisable, false)

	s.Reload(func(data *SafetyData) bool {
		data.Setter("new", "2.2.2.0/24", StatusEnable, false)
		return true
	})

	if got := s.Match("2.2.2.100", false); got != StatusEnable {
		t.Errorf("reload new rule: got %d, want %d", got, StatusEnable)
	}
	// 旧规则保留
	if got := s.Match("1.1.1.1", false); got != StatusDisable {
		t.Errorf("reload old rule: got %d, want %d", got, StatusDisable)
	}
}

func TestReloadAbort(t *testing.T) {
	s := New()
	s.Update("keep", "1.1.1.1", StatusEnable, false)

	s.Reload(func(data *SafetyData) bool {
		data.Setter("temp", "2.2.2.2", StatusDisable, false)
		return false // 不应用变更
	})

	if got := s.Match("2.2.2.2", false); got != StatusNone {
		t.Error("aborted reload should not add rule")
	}
}

func TestMatchInvalidIP(t *testing.T) {
	s := New()
	s.Update("any", "0.0.0.0~255.255.255.255", StatusDisable, false)

	if got := s.Match("", false); got != StatusNone {
		t.Error("empty IP should return StatusNone")
	}
	if got := s.Match("not-an-ip", false); got != StatusNone {
		t.Error("invalid IP should return StatusNone")
	}
}

func TestMatchWithPort(t *testing.T) {
	s := New()
	s.Update("web", "192.168.1.100", StatusEnable, false)

	// 带端口的 IP 应该剥离端口后匹配
	if got := s.Match("192.168.1.100:8080", false); got != StatusEnable {
		t.Errorf("IP with port: got %d, want %d", got, StatusEnable)
	}
}

func TestConcurrentMatchAndUpdate(t *testing.T) {
	s := New()
	for i := 0; i < 10; i++ {
		s.Update(fmt.Sprintf("rule_%d", i), fmt.Sprintf("10.0.%d.0/24", i), StatusEnable, false)
	}

	var wg sync.WaitGroup
	wg.Add(200)

	// 100 readers
	for i := 0; i < 100; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				s.Match(fmt.Sprintf("10.0.%d.%d", id%10, j%256), false)
			}
		}(i)
	}

	// 100 writers
	for i := 0; i < 100; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				name := fmt.Sprintf("dyn_%d_%d", id, j)
				s.Update(name, fmt.Sprintf("10.%d.%d.0/24", id%256, j%256), StatusDisable, false)
				s.Delete(name)
			}
		}(i)
	}

	wg.Wait()
}

// ==================== Benchmark ====================

func BenchmarkMatch_Hit(b *testing.B) {
	s := New()
	s.UseLocalAddress()
	for i := 0; i < 20; i++ {
		s.Update(fmt.Sprintf("rule_%d", i), fmt.Sprintf("10.%d.0.0/16", i), StatusEnable, false)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.Match("192.168.1.100", true)
	}
}

func BenchmarkMatch_Miss(b *testing.B) {
	s := New()
	for i := 0; i < 20; i++ {
		s.Update(fmt.Sprintf("rule_%d", i), fmt.Sprintf("10.%d.0.0/16", i), StatusEnable, false)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.Match("8.8.8.8", false)
	}
}

func BenchmarkParseIPv4(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		parseIPv4("192.168.1.100:8080")
	}
}

func BenchmarkMatch_Parallel(b *testing.B) {
	s := New()
	s.UseLocalAddress()
	for i := 0; i < 20; i++ {
		s.Update(fmt.Sprintf("rule_%d", i), fmt.Sprintf("10.%d.0.0/16", i), StatusDisable, false)
	}
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.Match("10.5.100.200", true)
		}
	})
}
