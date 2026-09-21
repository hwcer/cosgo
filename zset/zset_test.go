package zset

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// ==================== 功能测试 ====================

// 测试基本 ZAdd 功能
func TestZAdd(t *testing.T) {
	t.Log("\n========== 基本 ZAdd 测试 ==========")

	s := New()

	// 测试添加新元素
	score1 := s.ZAdd(100, "user1")
	if score1 != 100 {
		t.Errorf("ZAdd 失败，期望分数 100，实际分数 %d", score1)
	}

	// 测试更新已有元素的分数
	score2 := s.ZAdd(200, "user1")
	if score2 != 200 {
		t.Errorf("ZAdd 更新失败，期望分数 200，实际分数 %d", score2)
	}

	// 测试添加多个元素
	s.ZAdd(150, "user2")
	s.ZAdd(50, "user3")

	// 验证元素数量
	cardinality := s.ZCard()
	if cardinality != 3 {
		t.Errorf("ZCard 失败，期望数量 3，实际数量 %d", cardinality)
	}

	t.Log("基本 ZAdd 测试完成")
}

// 测试 ZRem 功能
func TestZRem(t *testing.T) {
	t.Log("\n========== ZRem 测试 ==========")

	s := New()

	// 添加元素
	s.ZAdd(100, "user1")
	s.ZAdd(200, "user2")
	s.ZAdd(300, "user3")

	// 测试删除不存在的元素
	removed1 := s.ZRem("user4")
	if removed1 {
		t.Error("ZRem 失败，删除不存在的元素应该返回 false")
	}

	// 测试删除存在的元素
	removed2 := s.ZRem("user2")
	if !removed2 {
		t.Error("ZRem 失败，删除存在的元素应该返回 true")
	}

	// 验证元素数量
	cardinality := s.ZCard()
	if cardinality != 2 {
		t.Errorf("ZCard 失败，期望数量 2，实际数量 %d", cardinality)
	}

	t.Log("ZRem 测试完成")
}

// 测试 ZRank 功能
func TestZRank(t *testing.T) {
	t.Log("\n========== ZRank 测试 ==========")

	s := New()

	// 添加元素
	s.ZAdd(100, "user1")
	s.ZAdd(200, "user2")
	s.ZAdd(300, "user3")

	// 测试获取排名
	rank1, score1 := s.ZRank("user1")
	if rank1 != 2 || score1 != 100 {
		t.Errorf("ZRank 失败，期望排名 2，分数 100，实际排名 %d，分数 %d", rank1, score1)
	}

	rank2, score2 := s.ZRank("user2")
	if rank2 != 1 || score2 != 200 {
		t.Errorf("ZRank 失败，期望排名 1，分数 200，实际排名 %d，分数 %d", rank2, score2)
	}

	rank3, score3 := s.ZRank("user3")
	if rank3 != 0 || score3 != 300 {
		t.Errorf("ZRank 失败，期望排名 0，分数 300，实际排名 %d，分数 %d", rank3, score3)
	}

	rank4, score4 := s.ZRank("user4")
	if rank4 != -1 || score4 != 0 {
		t.Errorf("ZRank 失败，期望排名 -1，分数 0，实际排名 %d，分数 %d", rank4, score4)
	}

	t.Log("ZRank 测试完成")
}

// 测试 ZScore 功能
func TestZScore(t *testing.T) {
	t.Log("\n========== ZScore 测试 ==========")

	s := New()

	// 添加元素
	s.ZAdd(100, "user1")
	s.ZAdd(200, "user2")

	// 测试获取分数
	score1, ok1 := s.ZScore("user1")
	if !ok1 || score1 != 100 {
		t.Errorf("ZScore 失败，期望分数 100，实际分数 %d，ok %v", score1, ok1)
	}

	score2, ok2 := s.ZScore("user2")
	if !ok2 || score2 != 200 {
		t.Errorf("ZScore 失败，期望分数 200，实际分数 %d，ok %v", score2, ok2)
	}

	score3, ok3 := s.ZScore("user3")
	if ok3 || score3 != 0 {
		t.Errorf("ZScore 失败，期望分数 0，ok false，实际分数 %d，ok %v", score3, ok3)
	}

	t.Log("ZScore 测试完成")
}

// 测试 ZElement 功能
func TestZElement(t *testing.T) {
	t.Log("\n========== ZElement 测试 ==========")

	s := New()

	// 添加元素
	s.ZAdd(100, "user1")
	s.ZAdd(200, "user2")
	s.ZAdd(300, "user3")

	// 测试获取指定排名的元素
	key1, score1 := s.ZElement(0)
	if key1 != "user3" || score1 != 300 {
		t.Errorf("ZElement 失败，期望 key user3，分数 300，实际 key %s，分数 %d", key1, score1)
	}

	key2, score2 := s.ZElement(1)
	if key2 != "user2" || score2 != 200 {
		t.Errorf("ZElement 失败，期望 key user2，分数 200，实际 key %s，分数 %d", key2, score2)
	}

	key3, score3 := s.ZElement(2)
	if key3 != "user1" || score3 != 100 {
		t.Errorf("ZElement 失败，期望 key user1，分数 100，实际 key %s，分数 %d", key3, score3)
	}

	key4, score4 := s.ZElement(3)
	if key4 != "" || score4 != 0 {
		t.Errorf("ZElement 失败，期望 key 空，分数 0，实际 key %s，分数 %d", key4, score4)
	}

	t.Log("ZElement 测试完成")
}

// 测试 ZRange 功能
func TestZRange(t *testing.T) {
	t.Log("\n========== ZRange 测试 ==========")

	s := New()

	// 添加元素
	s.ZAdd(100, "user1")
	s.ZAdd(200, "user2")
	s.ZAdd(300, "user3")
	s.ZAdd(400, "user4")

	// 测试范围遍历
	nodes := s.ZRange(0, 2)
	if len(nodes) != 3 {
		t.Errorf("ZRange 失败，期望 3 个元素，实际 %d 个元素", len(nodes))
	}

	expectedKeys := []string{"user4", "user3", "user2"}
	for i, node := range nodes {
		if node.Key != expectedKeys[i] {
			t.Errorf("ZRange 失败，期望第 %d 个元素 key %s，实际 key %s", i, expectedKeys[i], node.Key)
		}
	}

	t.Log("ZRange 测试完成")
}

// 测试 ZRangeByScore 功能
func TestZRangeByScore(t *testing.T) {
	t.Log("\n========== ZRangeByScore 测试 ==========")

	s := New()

	// 添加元素
	s.ZAdd(100, "user1")
	s.ZAdd(200, "user2")
	s.ZAdd(300, "user3")
	s.ZAdd(400, "user4")

	// 测试分数范围查询
	nodes := s.ZRangeByScore(200, 300)
	if len(nodes) != 2 {
		t.Errorf("ZRangeByScore 失败，期望 2 个元素，实际 %d 个元素", len(nodes))
	}

	expectedKeys := []string{"user3", "user2"}
	for i, node := range nodes {
		if node.Key != expectedKeys[i] {
			t.Errorf("ZRangeByScore 失败，期望第 %d 个元素 key %s，实际 key %s", i, expectedKeys[i], node.Key)
		}
	}

	t.Log("ZRangeByScore 测试完成")
}

// 测试 ZRemRangeByRank 功能
func TestZRemRangeByRank(t *testing.T) {
	t.Log("\n========== ZRemRangeByRank 测试 ==========")

	s := New()

	// 添加元素
	s.ZAdd(100, "user1")
	s.ZAdd(200, "user2")
	s.ZAdd(300, "user3")
	s.ZAdd(400, "user4")
	s.ZAdd(500, "user5")

	// 测试删除排名范围内的元素
	removed := s.ZRemRangeByRank(1, 3)
	if removed != 3 {
		t.Errorf("ZRemRangeByRank 失败，期望删除 3 个元素，实际删除 %d 个元素", removed)
	}

	// 验证剩余元素
	cardinality := s.ZCard()
	if cardinality != 2 {
		t.Errorf("ZCard 失败，期望数量 2，实际数量 %d", cardinality)
	}

	// 验证剩余元素是 user5 和 user1
	nodes := s.ZRange(0, 1)
	if len(nodes) != 2 {
		t.Errorf("ZRange 失败，期望 2 个元素，实际 %d 个元素", len(nodes))
	}

	expectedKeys := []string{"user5", "user1"}
	for i, node := range nodes {
		if node.Key != expectedKeys[i] {
			t.Errorf("ZRange 失败，期望第 %d 个元素 key %s，实际 key %s", i, expectedKeys[i], node.Key)
		}
	}

	t.Log("ZRemRangeByRank 测试完成")
}

// 测试 ZRemRangeByScore 功能
func TestZRemRangeByScore(t *testing.T) {
	t.Log("\n========== ZRemRangeByScore 测试 ==========")

	s := New()

	// 添加元素
	s.ZAdd(100, "user1")
	s.ZAdd(200, "user2")
	s.ZAdd(300, "user3")
	s.ZAdd(400, "user4")
	s.ZAdd(500, "user5")

	// 测试删除分数范围内的元素
	removed := s.ZRemRangeByScore(200, 400)
	if removed != 3 {
		t.Errorf("ZRemRangeByScore 失败，期望删除 3 个元素，实际删除 %d 个元素", removed)
	}

	// 验证剩余元素
	cardinality := s.ZCard()
	if cardinality != 2 {
		t.Errorf("ZCard 失败，期望数量 2，实际数量 %d", cardinality)
	}

	// 验证剩余元素是 user5 和 user1
	nodes := s.ZRange(0, 1)
	if len(nodes) != 2 {
		t.Errorf("ZRange 失败，期望 2 个元素，实际 %d 个元素", len(nodes))
	}

	expectedKeys := []string{"user5", "user1"}
	for i, node := range nodes {
		if node.Key != expectedKeys[i] {
			t.Errorf("ZRange 失败，期望第 %d 个元素 key %s，实际 key %s", i, expectedKeys[i], node.Key)
		}
	}

	t.Log("ZRemRangeByScore 测试完成")
}

// 测试 ZCount 功能
func TestZCount(t *testing.T) {
	t.Log("\n========== ZCount 测试 ==========")

	s := New()

	// 添加元素
	s.ZAdd(100, "user1")
	s.ZAdd(200, "user2")
	s.ZAdd(300, "user3")
	s.ZAdd(400, "user4")
	s.ZAdd(500, "user5")

	// 测试分数范围内的元素数量
	count := s.ZCount(200, 400)
	if count != 3 {
		t.Errorf("ZCount 失败，期望数量 3，实际数量 %d", count)
	}

	count = s.ZCount(100, 500)
	if count != 5 {
		t.Errorf("ZCount 失败，期望数量 5，实际数量 %d", count)
	}

	count = s.ZCount(600, 700)
	if count != 0 {
		t.Errorf("ZCount 失败，期望数量 0，实际数量 %d", count)
	}

	t.Log("ZCount 测试完成")
}

// 测试 ZIncr 功能
func TestZIncr(t *testing.T) {
	t.Log("\n========== ZIncr 测试 ==========")

	s := New()

	// 添加元素
	s.ZAdd(100, "user1")

	// 测试增加分数
	newScore := s.ZIncr(50, "user1")
	if newScore != 150 {
		t.Errorf("ZIncr 失败，期望分数 150，实际分数 %d", newScore)
	}

	// 测试减少分数
	newScore = s.ZIncr(-30, "user1")
	if newScore != 120 {
		t.Errorf("ZIncr 失败，期望分数 120，实际分数 %d", newScore)
	}

	// 测试对不存在的元素增加分数
	newScore = s.ZIncr(100, "user2")
	if newScore != 100 {
		t.Errorf("ZIncr 失败，期望分数 100，实际分数 %d", newScore)
	}

	t.Log("ZIncr 测试完成")
}

// 测试同分先到先得（FIFO）
func TestSameScoreFIFO(t *testing.T) {
	t.Log("\n========== 同分FIFO测试 ==========")

	// 降序模式
	s := New(-1)
	s.ZAdd(100, "first")
	s.ZAdd(100, "second")
	s.ZAdd(100, "third")

	// 先到的排名应该靠前
	rank1, _ := s.ZRank("first")
	rank2, _ := s.ZRank("second")
	rank3, _ := s.ZRank("third")
	if rank1 != 0 || rank2 != 1 || rank3 != 2 {
		t.Errorf("FIFO 失败(降序)，期望 first=0,second=1,third=2，实际 %d,%d,%d", rank1, rank2, rank3)
	}

	// ZRange 顺序应该与插入顺序一致
	nodes := s.ZRange(0, 2)
	expected := []string{"first", "second", "third"}
	for i, node := range nodes {
		if node.Key != expected[i] {
			t.Errorf("ZRange FIFO 失败，期望第 %d 个 %s，实际 %s", i, expected[i], node.Key)
		}
	}

	// 删除中间元素后排名仍正确
	s.ZRem("second")
	rank1, _ = s.ZRank("first")
	rank3, _ = s.ZRank("third")
	if rank1 != 0 || rank3 != 1 {
		t.Errorf("FIFO 删除后失败，期望 first=0,third=1，实际 %d,%d", rank1, rank3)
	}

	// 升序模式
	s2 := New(1)
	s2.ZAdd(50, "alpha")
	s2.ZAdd(50, "beta")
	s2.ZAdd(50, "gamma")

	r1, _ := s2.ZRank("alpha")
	r2, _ := s2.ZRank("beta")
	r3, _ := s2.ZRank("gamma")
	if r1 != 0 || r2 != 1 || r3 != 2 {
		t.Errorf("FIFO 失败(升序)，期望 alpha=0,beta=1,gamma=2，实际 %d,%d,%d", r1, r2, r3)
	}

	t.Log("同分FIFO测试完成")
}

// 测试守门员同分拒绝
func TestGuardSameScoreReject(t *testing.T) {
	t.Log("\n========== 守门员同分拒绝测试 ==========")

	s := NewWithMaxSize(3)
	s.ZAdd(300, "a")
	s.ZAdd(200, "b")
	s.ZAdd(100, "c") // 守门员：100, "c"

	// 新成员分数等于守门员，应被拒绝（先到者优先）
	result := s.ZAdd(100, "d")
	if result != 0 {
		t.Errorf("期望同分新成员被拒绝(返回0)，实际返回 %d", result)
	}

	// 新成员分数低于守门员，应被拒绝
	result = s.ZAdd(50, "e")
	if result != 0 {
		t.Errorf("期望低分新成员被拒绝(返回0)，实际返回 %d", result)
	}

	// 新成员分数高于守门员，应被接受
	result = s.ZAdd(150, "f")
	if result != 150 {
		t.Errorf("期望高分新成员被接受(返回150)，实际返回 %d", result)
	}

	t.Log("守门员同分拒绝测试完成")
}

// 测试倒序排列
func TestZSetDescending(t *testing.T) {
	t.Log("\n========== 倒序排列测试 ==========")

	s := New(-1) // -1 表示倒序

	// 添加元素
	s.ZAdd(100, "user1")
	s.ZAdd(200, "user2")
	s.ZAdd(300, "user3")

	// 验证顺序是倒序（分数从高到低）
	nodes := s.ZRange(0, 2)
	if len(nodes) != 3 {
		t.Errorf("ZRange 失败，期望 3 个元素，实际 %d 个元素", len(nodes))
	}

	expectedKeys := []string{"user3", "user2", "user1"}
	for i, node := range nodes {
		if node.Key != expectedKeys[i] {
			t.Errorf("ZRange 失败，期望第 %d 个元素 key %s，实际 key %s", i, expectedKeys[i], node.Key)
		}
	}

	t.Log("倒序排列测试完成")
}

// 测试正序排列
func TestZSetAscending(t *testing.T) {
	t.Log("\n========== 正序排列测试 ==========")

	s := New(1) // 1 表示正序

	// 添加元素
	s.ZAdd(100, "user1")
	s.ZAdd(200, "user2")
	s.ZAdd(300, "user3")

	// 验证顺序是正序（分数从低到高）
	nodes := s.ZRange(0, 2)
	if len(nodes) != 3 {
		t.Errorf("ZRange 失败，期望 3 个元素，实际 %d 个元素", len(nodes))
	}

	expectedKeys := []string{"user1", "user2", "user3"}
	for i, node := range nodes {
		if node.Key != expectedKeys[i] {
			t.Errorf("ZRange 失败，期望第 %d 个元素 key %s，实际 key %s", i, expectedKeys[i], node.Key)
		}
	}

	t.Log("正序排列测试完成")
}

// 测试 MAXSIZE 功能
func TestZSetMaxSize(t *testing.T) {
	t.Log("\n========== MAXSIZE 测试 ==========")

	// 创建带人数限制的 ZSet
	s := NewWithMaxSize(3)

	// 添加超过限制的元素
	s.ZAdd(100, "user1")
	s.ZAdd(200, "user2")
	s.ZAdd(300, "user3")
	s.ZAdd(400, "user4") // 这会导致 user1 被删除

	// 验证元素数量
	cardinality := s.ZCard()
	if cardinality != 3 {
		t.Errorf("ZCard 失败，期望数量 3，实际数量 %d", cardinality)
	}

	// 验证 user1 已被删除
	rank, _ := s.ZRank("user1")
	if rank != -1 {
		t.Errorf("ZRank 失败，user1 应该被删除，实际排名 %d", rank)
	}

	// 验证 user2、user3、user4 存在
	rank, _ = s.ZRank("user2")
	if rank == -1 {
		t.Error("ZRank 失败，user2 应该存在")
	}

	rank, _ = s.ZRank("user3")
	if rank == -1 {
		t.Error("ZRank 失败，user3 应该存在")
	}

	rank, _ = s.ZRank("user4")
	if rank == -1 {
		t.Error("ZRank 失败，user4 应该存在")
	}

	t.Log("MAXSIZE 测试完成")
}

// 测试裁剪冗余功能
func TestTrimExcess(t *testing.T) {
	t.Log("\n========== 裁剪冗余测试 ==========")

	// 创建带人数限制的 ZSet
	s := NewWithMaxSize(5)

	// 添加超过限制的元素
	for i := range 20 {
		s.ZAdd(int64(20-i), fmt.Sprintf("user%d", i))
	}

	// 手动触发裁剪
	s.tryTrimExcess()

	// 验证元素数量
	cardinality := s.ZCard()
	if cardinality > 5 {
		t.Errorf("tryTrimExcess 失败，期望数量 <= 5，实际数量 %d", cardinality)
	}

	t.Log("裁剪后测试完成")
}

// ==================== 线程安全测试 ====================

// 测试 ZSet 线程安全
func TestZSetThreadSafe(t *testing.T) {
	t.Log("\n========== ZSet 线程安全测试 ==========")

	s := New()

	const goroutines = 100
	const operations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	start := time.Now()

	for i := range goroutines {
		go func(id int) {
			defer wg.Done()
			for j := range operations {
				key := fmt.Sprintf("user_%d_%d", id, j)
				score := int64(id*10000 + j)
				s.ZAdd(score, key)
				if j%10 == 0 {
					s.ZRem(key)
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	cardinality := s.ZCard()
	t.Logf("线程安全测试完成，耗时 %v，最终元素数量 %d", duration, cardinality)
}

// 测试 ZSet 带 MAXSIZE 线程安全
func TestZSetWithMaxSizeThreadSafe(t *testing.T) {
	t.Log("\n========== ZSet 带 MAXSIZE 线程安全测试 ==========")

	s := NewWithMaxSize(1000)

	const goroutines = 100
	const operations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	start := time.Now()

	for i := range goroutines {
		go func(id int) {
			defer wg.Done()
			for j := range operations {
				key := fmt.Sprintf("user_%d", id)
				score := int64(id*10000 + j)
				s.ZAdd(score, key)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	cardinality := s.ZCard()
	t.Logf("线程安全测试完成，耗时 %v，最终元素数量 %d", duration, cardinality)
}

// 测试满员时榜内成员降分后守门员必须重算,否则CanEnter出现假阴性、ZRank误返回-1
func TestGuardUpdateOnMemberScoreDrop(t *testing.T) {
	s := NewWithMaxSize(3)
	s.ZAdd(10, "A")
	s.ZAdd(8, "B")
	s.ZAdd(5, "C")

	// B从8降到3,实际守门员应从C(5)变为B(3)
	if r := s.ZAdd(3, "B"); r != 3 {
		t.Fatalf("ZAdd 降分失败，期望 3，实际 %d", r)
	}

	// 此刻榜内为[A10,C5,B3],B仍在榜内,ZRank不应返回-1
	if rank, _ := s.ZRank("B"); rank != 2 {
		t.Errorf("ZRank 误判：B 在榜内排名 2，实际 %d", rank)
	}

	// F(4)优于实际守门员B(3)，应能入榜
	if !s.CanEnter(4) {
		t.Error("CanEnter 假阴性：4 应优于守门员 3")
	}
	if r := s.ZAdd(4, "F"); r != 4 {
		t.Errorf("ZAdd 被陈旧守门员拦截，期望 4，实际 %d", r)
	}

	// F入榜后为[A10,C5,F4,B3],B跌出前3,返回-1是正确语义
	if rank, _ := s.ZRank("F"); rank != 2 {
		t.Errorf("ZRank 误判：F 排名 2，实际 %d", rank)
	}
	if rank, _ := s.ZRank("B"); rank != -1 {
		t.Errorf("ZRank 误判：B 已跌出榜外，期望 -1，实际 %d", rank)
	}
}

// 回归：删除与守门员同分的榜内成员后必须重算守门员。
// 漏重算时守门员陈旧，导致新成员被误拒（丢写入）、CanEnter 假阴性。
func TestZRemSameScoreTieUpdatesGuard(t *testing.T) {
	s := NewWithMaxSize(3, 1) // 升序
	s.ZAdd(20, "A")
	s.ZAdd(20, "B")
	s.ZAdd(30, "C") // 满员 A20 B20 C30，守门员 C30
	s.ZAdd(10, "E") // 入榜 → E10 A20 B20 C30(冗余)，守门员 B20

	// A 与守门员 B20 同分，删除后冗余成员 C30 顶位，真实守门员应为 C30
	if !s.ZRem("A") {
		t.Fatal("ZRem(A) 失败")
	}

	if !s.CanEnter(25) {
		t.Error("CanEnter 假阴性：真实守门员 C30，25 应可入榜")
	}
	if r := s.ZAdd(25, "F"); r != 25 {
		t.Errorf("ZAdd 被陈旧守门员拦截，期望 25，实际 %d", r)
	}
}

// 回归：与守门员同分的冗余成员 ZRank 必须返回 -1，
// 泄漏真实排名会造成与 ZCard 的矛盾（rank >= maxSize）
func TestZRankGuardTieBoundary(t *testing.T) {
	s := NewWithMaxSize(3)
	s.ZAdd(100, "A")
	s.ZAdd(90, "B")
	s.ZAdd(90, "C") // 满员，守门员 C90
	s.ZAdd(95, "E") // 入榜 → A100 E95 B90 C90(冗余)，守门员 B90

	if rank, _ := s.ZRank("C"); rank != -1 {
		t.Errorf("冗余成员同分边界：期望 ZRank(C)=-1，实际 %d（ZCard=%d）", rank, s.ZCard())
	}
	// 榜内同分成员不受影响
	if rank, _ := s.ZRank("B"); rank != 2 {
		t.Errorf("榜内成员：期望 ZRank(B)=2，实际 %d", rank)
	}
}

// ZIncr(0) 对不存在的 key：无守门员时创建 0 分成员，有守门员时不创建
func TestZIncrZeroRebuild(t *testing.T) {
	s := New()
	if r := s.ZIncr(0, "new"); r != 0 {
		t.Errorf("期望返回 0，实际 %d", r)
	}
	if score, ok := s.ZScore("new"); !ok || score != 0 {
		t.Errorf("无守门员时 ZIncr(0) 应创建成员，ZScore=(%d,%v)", score, ok)
	}
	if n := s.ZCard(); n != 1 {
		t.Errorf("无守门员时成员数期望 1，实际 %d", n)
	}
	// 已存在成员仍是纯读
	if r := s.ZIncr(0, "new"); r != 0 {
		t.Errorf("已存在成员 ZIncr(0) 期望返回当前分 0，实际 %d", r)
	}

	g := NewWithMaxSize(3)
	g.ZAdd(10, "a")
	g.ZAdd(20, "b")
	g.ZAdd(30, "c")
	if r := g.ZIncr(0, "new"); r != 0 {
		t.Errorf("期望返回 0，实际 %d", r)
	}
	if _, ok := g.ZScore("new"); ok {
		t.Error("有守门员时 ZIncr(0) 不应创建成员")
	}
	if n := g.ZCard(); n != 3 {
		t.Errorf("有守门员时成员数不应变化，期望 3，实际 %d", n)
	}
}
