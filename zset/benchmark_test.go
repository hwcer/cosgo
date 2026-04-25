package zset

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// 综合基准测试：10000 用户并发读写，排行榜上限 2000
// =============================================================================

const (
	benchUsers   = 10000
	benchMaxSize = 2000
)

// TestBenchmarkMixedReadWrite 综合读写混合测试
// 模拟真实排行榜场景：10000 用户同时进行分数更新 + 排名查询 + 范围查询
func TestBenchmarkMixedReadWrite(t *testing.T) {
	set := NewWithMaxSize(benchMaxSize)

	// 预热：先填充排行榜
	for i := 0; i < benchUsers; i++ {
		set.ZAdd(int64(rand.Intn(100000)), fmt.Sprintf("user_%d", i))
	}

	const rounds = 50 // 每用户操作轮次
	var totalOps atomic.Int64
	var writeOps atomic.Int64
	var readOps atomic.Int64

	var wg sync.WaitGroup
	wg.Add(benchUsers)
	start := time.Now()

	for i := 0; i < benchUsers; i++ {
		go func(uid int) {
			defer wg.Done()
			key := fmt.Sprintf("user_%d", uid)
			rng := rand.New(rand.NewSource(int64(uid)))

			for r := 0; r < rounds; r++ {
				op := rng.Intn(100)
				switch {
				case op < 40:
					// 40% ZAdd（更新分数）
					set.ZAdd(int64(rng.Intn(100000)), key)
					writeOps.Add(1)
				case op < 55:
					// 15% ZIncr（增量更新）
					set.ZIncr(int64(rng.Intn(100)-50), key)
					writeOps.Add(1)
				case op < 75:
					// 20% ZRank（查排名）
					set.ZRank(key)
					readOps.Add(1)
				case op < 85:
					// 10% ZScore（查分数）
					set.ZScore(key)
					readOps.Add(1)
				case op < 95:
					// 10% ZRange（Top 10）
					set.ZRange(0, 9)
					readOps.Add(1)
				default:
					// 5% ZElement（随机排名查询）
					set.ZElement(int64(rng.Intn(benchMaxSize)))
					readOps.Add(1)
				}
				totalOps.Add(1)
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	total := totalOps.Load()
	writes := writeOps.Load()
	reads := readOps.Load()
	ops := float64(total) / elapsed.Seconds()

	t.Logf("========== 混合读写测试 ==========")
	t.Logf("用户数: %d, 排行榜上限: %d, 每用户轮次: %d", benchUsers, benchMaxSize, rounds)
	t.Logf("总操作: %d (写 %d / 读 %d)", total, writes, reads)
	t.Logf("总耗时: %v", elapsed)
	t.Logf("吞吐量: %.0f ops/sec", ops)
	t.Logf("平均延迟: %.2f µs/op", float64(elapsed.Microseconds())/float64(total))
}

// TestBenchmarkPureWrite 纯写入压测
func TestBenchmarkPureWrite(t *testing.T) {
	set := NewWithMaxSize(benchMaxSize)
	var totalOps atomic.Int64

	var wg sync.WaitGroup
	wg.Add(benchUsers)
	start := time.Now()

	for i := 0; i < benchUsers; i++ {
		go func(uid int) {
			defer wg.Done()
			key := fmt.Sprintf("user_%d", uid)
			for r := 0; r < 100; r++ {
				set.ZAdd(int64(rand.Intn(100000)), key)
				totalOps.Add(1)
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	total := totalOps.Load()

	t.Logf("========== 纯写入测试 ==========")
	t.Logf("用户数: %d, 排行榜上限: %d, 每用户写入: 100 次", benchUsers, benchMaxSize)
	t.Logf("总写入: %d", total)
	t.Logf("总耗时: %v", elapsed)
	t.Logf("吞吐量: %.0f ops/sec", float64(total)/elapsed.Seconds())
	t.Logf("平均延迟: %.2f µs/op", float64(elapsed.Microseconds())/float64(total))
}

// TestBenchmarkPureRead 纯读取压测（预填充后并发读）
func TestBenchmarkPureRead(t *testing.T) {
	set := NewWithMaxSize(benchMaxSize)
	for i := 0; i < benchUsers; i++ {
		set.ZAdd(int64(rand.Intn(100000)), fmt.Sprintf("user_%d", i))
	}

	var totalOps atomic.Int64

	var wg sync.WaitGroup
	wg.Add(benchUsers)
	start := time.Now()

	for i := 0; i < benchUsers; i++ {
		go func(uid int) {
			defer wg.Done()
			key := fmt.Sprintf("user_%d", uid)
			rng := rand.New(rand.NewSource(int64(uid)))
			for r := 0; r < 100; r++ {
				switch rng.Intn(4) {
				case 0:
					set.ZRank(key)
				case 1:
					set.ZScore(key)
				case 2:
					set.ZRange(0, 9)
				case 3:
					set.ZElement(int64(rng.Intn(benchMaxSize)))
				}
				totalOps.Add(1)
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	total := totalOps.Load()

	t.Logf("========== 纯读取测试 ==========")
	t.Logf("用户数: %d, 排行榜上限: %d, 每用户读取: 100 次", benchUsers, benchMaxSize)
	t.Logf("总读取: %d", total)
	t.Logf("总耗时: %v", elapsed)
	t.Logf("吞吐量: %.0f ops/sec", float64(total)/elapsed.Seconds())
	t.Logf("平均延迟: %.2f µs/op", float64(elapsed.Microseconds())/float64(total))
}

// TestBenchmarkHighContention 高争用测试（少量热点 key 被大量协程同时写）
func TestBenchmarkHighContention(t *testing.T) {
	set := NewWithMaxSize(benchMaxSize)
	for i := 0; i < benchUsers; i++ {
		set.ZAdd(int64(i), fmt.Sprintf("user_%d", i))
	}

	const hotKeys = 100 // 100 个热点 key 被 10000 协程争抢
	var totalOps atomic.Int64

	var wg sync.WaitGroup
	wg.Add(benchUsers)
	start := time.Now()

	for i := 0; i < benchUsers; i++ {
		go func(uid int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(int64(uid)))
			for r := 0; r < 100; r++ {
				hotKey := fmt.Sprintf("user_%d", rng.Intn(hotKeys))
				set.ZIncr(int64(rng.Intn(10)+1), hotKey)
				totalOps.Add(1)
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	total := totalOps.Load()

	t.Logf("========== 高争用测试 ==========")
	t.Logf("用户数: %d, 热点 key: %d, 排行榜上限: %d", benchUsers, hotKeys, benchMaxSize)
	t.Logf("总操作: %d", total)
	t.Logf("总耗时: %v", elapsed)
	t.Logf("吞吐量: %.0f ops/sec", float64(total)/elapsed.Seconds())
	t.Logf("平均延迟: %.2f µs/op", float64(elapsed.Microseconds())/float64(total))
}

// TestBenchmarkGuardReject 守门员拦截效率测试
// 大量低分写入被守门员快速拒绝的场景
func TestBenchmarkGuardReject(t *testing.T) {
	set := NewWithMaxSize(benchMaxSize)

	// 预填充高分数据，使守门员分数很高
	for i := 0; i < benchMaxSize; i++ {
		set.ZAdd(int64(50000+i), fmt.Sprintf("top_%d", i))
	}

	var accepted atomic.Int64
	var rejected atomic.Int64

	var wg sync.WaitGroup
	wg.Add(benchUsers)
	start := time.Now()

	for i := 0; i < benchUsers; i++ {
		go func(uid int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(int64(uid)))
			for r := 0; r < 100; r++ {
				key := fmt.Sprintf("new_%d_%d", uid, r)
				// 80% 低分（被拦截），20% 高分（可能入榜）
				var score int64
				if rng.Intn(100) < 80 {
					score = int64(rng.Intn(50000))
				} else {
					score = int64(50000 + rng.Intn(50000))
				}
				result := set.ZAdd(score, key)
				if result == 0 {
					rejected.Add(1)
				} else {
					accepted.Add(1)
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	total := accepted.Load() + rejected.Load()

	t.Logf("========== 守门员拦截测试 ==========")
	t.Logf("用户数: %d, 排行榜上限: %d", benchUsers, benchMaxSize)
	t.Logf("总操作: %d (入榜 %d / 拒绝 %d)", total, accepted.Load(), rejected.Load())
	t.Logf("拒绝率: %.1f%%", float64(rejected.Load())/float64(total)*100)
	t.Logf("总耗时: %v", elapsed)
	t.Logf("吞吐量: %.0f ops/sec", float64(total)/elapsed.Seconds())
	t.Logf("平均延迟: %.2f µs/op", float64(elapsed.Microseconds())/float64(total))
}

// TestBenchmarkLatencyDistribution 延迟分布测试
// 单独测量各操作的 P50/P99 延迟
func TestBenchmarkLatencyDistribution(t *testing.T) {
	set := NewWithMaxSize(benchMaxSize)
	for i := 0; i < benchUsers; i++ {
		set.ZAdd(int64(rand.Intn(100000)), fmt.Sprintf("user_%d", i))
	}

	type opResult struct {
		name     string
		samples  []time.Duration
	}

	ops := []struct {
		name string
		fn   func()
	}{
		{"ZAdd", func() { set.ZAdd(int64(rand.Intn(100000)), fmt.Sprintf("user_%d", rand.Intn(benchUsers))) }},
		{"ZIncr", func() { set.ZIncr(int64(rand.Intn(100)), fmt.Sprintf("user_%d", rand.Intn(benchUsers))) }},
		{"ZRank", func() { set.ZRank(fmt.Sprintf("user_%d", rand.Intn(benchUsers))) }},
		{"ZScore", func() { set.ZScore(fmt.Sprintf("user_%d", rand.Intn(benchUsers))) }},
		{"ZRange(Top10)", func() { set.ZRange(0, 9) }},
		{"ZRange(Top100)", func() { set.ZRange(0, 99) }},
		{"ZElement", func() { set.ZElement(int64(rand.Intn(benchMaxSize))) }},
		{"ZCard", func() { set.ZCard() }},
		{"ZCount", func() { set.ZCount(10000, 90000) }},
		{"ZRangeByScore", func() { set.ZRangeByScore(40000, 60000) }},
	}

	const samplesPerOp = 10000

	t.Logf("========== 单操作延迟测试（单线程，%d 次采样）==========", samplesPerOp)
	for _, op := range ops {
		samples := make([]time.Duration, samplesPerOp)
		for i := 0; i < samplesPerOp; i++ {
			s := time.Now()
			op.fn()
			samples[i] = time.Since(s)
		}

		// 排序取百分位
		sortDurations(samples)
		p50 := samples[samplesPerOp*50/100]
		p95 := samples[samplesPerOp*95/100]
		p99 := samples[samplesPerOp*99/100]

		var total time.Duration
		for _, d := range samples {
			total += d
		}
		avg := total / time.Duration(samplesPerOp)

		t.Logf("%-18s  avg=%-8v  P50=%-8v  P95=%-8v  P99=%-8v", op.name, avg, p50, p95, p99)
	}
}

// sortDurations 插入排序（样本量不大，足够用）
func sortDurations(d []time.Duration) {
	for i := 1; i < len(d); i++ {
		key := d[i]
		j := i - 1
		for j >= 0 && d[j] > key {
			d[j+1] = d[j]
			j--
		}
		d[j+1] = key
	}
}

// =============================================================================
// Go 官方 Benchmark（go test -bench=. 使用）
// =============================================================================

func BenchmarkZAdd_NoMaxSize(b *testing.B) {
	set := New()
	for i := 0; i < 10000; i++ {
		set.ZAdd(int64(i), fmt.Sprintf("user_%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.ZAdd(int64(rand.Intn(100000)), fmt.Sprintf("user_%d", rand.Intn(10000)))
	}
}

func BenchmarkZAdd_WithMaxSize(b *testing.B) {
	set := NewWithMaxSize(benchMaxSize)
	for i := 0; i < 10000; i++ {
		set.ZAdd(int64(i), fmt.Sprintf("user_%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.ZAdd(int64(rand.Intn(100000)), fmt.Sprintf("user_%d", rand.Intn(10000)))
	}
}

func BenchmarkZIncr(b *testing.B) {
	set := NewWithMaxSize(benchMaxSize)
	for i := 0; i < 10000; i++ {
		set.ZAdd(int64(i), fmt.Sprintf("user_%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.ZIncr(int64(rand.Intn(100)), fmt.Sprintf("user_%d", rand.Intn(10000)))
	}
}

func BenchmarkZRank(b *testing.B) {
	set := NewWithMaxSize(benchMaxSize)
	for i := 0; i < 10000; i++ {
		set.ZAdd(int64(rand.Intn(100000)), fmt.Sprintf("user_%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.ZRank(fmt.Sprintf("user_%d", rand.Intn(10000)))
	}
}

func BenchmarkZScore(b *testing.B) {
	set := NewWithMaxSize(benchMaxSize)
	for i := 0; i < 10000; i++ {
		set.ZAdd(int64(rand.Intn(100000)), fmt.Sprintf("user_%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.ZScore(fmt.Sprintf("user_%d", rand.Intn(10000)))
	}
}

func BenchmarkZRange_Top10(b *testing.B) {
	set := NewWithMaxSize(benchMaxSize)
	for i := 0; i < 10000; i++ {
		set.ZAdd(int64(rand.Intn(100000)), fmt.Sprintf("user_%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.ZRange(0, 9)
	}
}

func BenchmarkZRange_Top100(b *testing.B) {
	set := NewWithMaxSize(benchMaxSize)
	for i := 0; i < 10000; i++ {
		set.ZAdd(int64(rand.Intn(100000)), fmt.Sprintf("user_%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.ZRange(0, 99)
	}
}

func BenchmarkZElement(b *testing.B) {
	set := NewWithMaxSize(benchMaxSize)
	for i := 0; i < 10000; i++ {
		set.ZAdd(int64(rand.Intn(100000)), fmt.Sprintf("user_%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.ZElement(int64(rand.Intn(benchMaxSize)))
	}
}

func BenchmarkParallel_MixedReadWrite(b *testing.B) {
	set := NewWithMaxSize(benchMaxSize)
	for i := 0; i < 10000; i++ {
		set.ZAdd(int64(rand.Intn(100000)), fmt.Sprintf("user_%d", i))
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			uid := rng.Intn(10000)
			key := fmt.Sprintf("user_%d", uid)
			switch rng.Intn(5) {
			case 0:
				set.ZAdd(int64(rng.Intn(100000)), key)
			case 1:
				set.ZIncr(int64(rng.Intn(100)), key)
			case 2:
				set.ZRank(key)
			case 3:
				set.ZRange(0, 9)
			case 4:
				set.ZScore(key)
			}
		}
	})
}
