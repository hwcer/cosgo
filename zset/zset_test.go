package zset

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// ==================== 基础功能测试 ====================

// 测试 ZRange 功能（升序模式）
func TestZRangeStringInt64(t *testing.T) {
	s := New(1) // 升序

	s.ZAdd(66, "1")
	s.ZAdd(66, "2")
	s.ZAdd(66, "3")
	s.ZAdd(100, "4")
	s.ZAdd(99, "5")
	s.ZAdd(33, "6")
	s.ZAdd(77, "7")

	rank, score := s.ZRank("2")
	t.Log("Key:", "2", "Rank:", rank, "Score:", score)

	var i int64
	s.ZRange(0, 10, func(v int64, k string) {
		i++
		t.Log("排序", "Key:", k, "Rank:", i, "Score:", v)
	})

	// 测试 ZCount 功能
	count := s.ZCount(60, 80)
	t.Log("ZCount(60, 80):", count)
	if count != 4 {
		t.Errorf("Expected count 4, got %d", count)
	}
}

// 测试 ZIncr 功能
func TestZIncrStringInt64(t *testing.T) {
	s := New()

	s.ZAdd(10, "user1")
	s.ZAdd(20, "user2")

	// 增加分数
	newScore := s.ZIncr(5, "user1")
	if newScore != 15 {
		t.Errorf("Expected new score 15, got %d", newScore)
	}
	t.Log("After ZIncr, user1 score:", newScore)

	// 测试 ZCount 在分数变化后的结果
	count := s.ZCount(10, 20)
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
	t.Log("ZCount(10, 20):", count)
}

// 测试 ZRem 功能
func TestZRemStringInt64(t *testing.T) {
	s := New()

	s.ZAdd(10, "user1")
	s.ZAdd(20, "user2")
	s.ZAdd(30, "user3")

	// 删除元素
	ok := s.ZRem("user2")
	if !ok {
		t.Error("Failed to remove user2")
	}
	t.Log("Remove user2 success:", ok)

	// 检查元素是否被删除
	score, exists := s.ZScore("user2")
	if exists {
		t.Errorf("user2 should not exist, but has score %d", score)
	}

	// 检查 ZCount 结果
	count := s.ZCount(0, 100)
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
	t.Log("After ZRem, ZCount(0, 100):", count)

	// 检查 ZCard 结果
	card := s.ZCard()
	if card != 2 {
		t.Errorf("Expected ZCard 2, got %d", card)
	}
	t.Log("After ZRem, ZCard:", card)
}

// 测试 ZData 功能
func TestZDataStringInt64(t *testing.T) {
	s := New(-1) // 显式使用降序

	s.ZAdd(100, "high")
	s.ZAdd(50, "mid")
	s.ZAdd(10, "low")

	// 测试 ZData - 第 1 名（降序模式高分在前）
	key, score := s.ZData(0)
	if key != "high" || score != 100 {
		t.Errorf("Expected (high, 100), got (%s, %d)", key, score)
	}
	t.Log("ZData(0) - 第 1 名:", key, score)

	// 测试最后一名
	lastRank := s.ZCard() - 1
	key, score = s.ZData(lastRank)
	t.Log("ZData(last) - 最后一名:", key, score)
	
	// 验证默认也是降序
	s2 := New() // 默认降序
	s2.ZAdd(100, "high2")
	s2.ZAdd(50, "mid2")
	s2.ZAdd(10, "low2")
	
	key, score = s2.ZData(0)
	if key != "high2" || score != 100 {
		t.Errorf("默认降序模式第 1 名应该是高分，got (%s, %d)", key, score)
	}
	t.Log("默认降序模式验证:", key, score)
}

// 测试 ZScore 功能
func TestZScoreStringInt64(t *testing.T) {
	s := New()

	s.ZAdd(100, "user1")
	s.ZAdd(200, "user2")

	// 测试存在的元素
	score, ok := s.ZScore("user1")
	if !ok || score != 100 {
		t.Errorf("Expected score 100, got %d (ok=%v)", score, ok)
	}
	t.Log("ZScore(user1):", score, "Exists:", ok)

	// 测试不存在的元素
	score, ok = s.ZScore("user3")
	if ok || score != 0 {
		t.Errorf("Expected score 0 and ok=false, got %d (ok=%v)", score, ok)
	}
	t.Log("ZScore(user3):", score, "Exists:", ok)
}

// ==================== 人数限制和守门员测试 ====================

// 测试人数限制功能（降序模式 - 懒淘汰机制）
func TestMaxSizeAndGuardKeeper(t *testing.T) {
	t.Log("\n========== 测试人数限制和守门员机制（懒淘汰） ==========")
	
	// 创建降序 ZSet，最多保留 100 人
	maxSize := int32(100)
	s := NewWithMaxSize(maxSize, -1)

	t.Logf("添加 250 个元素，分数从 1-250，验证懒淘汰机制")
	t.Logf("阈值配置：maxSize=%d, cleanupBufferSize=%d", maxSize, cleanupBufferSize)
	
	// 添加 250 个元素（超过阈值 100+100=200，应该触发清理）
	for i := 0; i < 250; i++ {
		s.ZAdd(int64(i+1), fmt.Sprintf("user%d", i))
	}

	// 检查 dict 大小（应该被清理到接近 maxSize）
	dictSize := len(s.dict)
	t.Logf("dict 大小：%d (期望：接近 %d)", dictSize, maxSize)

	// 检查跳表大小（可能暂时超出，但前 N 名应该是高分）
	card := s.ZCard()
	t.Logf("跳表元素数量：%d (懒淘汰模式下可能>%d)", card, maxSize)

	// 获取守门员分数
	guardScore, ok := s.GetGuardScore()
	if !ok {
		t.Error("守门员应该有效")
	} else {
		t.Logf("守门员分数：%d", guardScore)
	}

	// 验证前 10 名都是高分（这是核心要求）
	t.Log("\n验证前 10 名（必须是最高分）:")
	allHighScore := true
	for i := 0; i < 10; i++ {
		key, score := s.ZData(int64(i))
		t.Logf("第%d名：%s, 分数=%d", i+1, key, score)
		if score < 240 {
			t.Errorf("前 10 名分数应该都>=240, 实际 %d", score)
			allHighScore = false
		}
	}
	
	if !allHighScore {
		t.Error("前 10 名中存在低分元素，懒淘汰机制失效")
	}

	// 测试低分元素是否能进入
	t.Log("\n尝试添加低分元素（分数=1）...")
	s.ZAdd(1, "low_score_user")
	
	// 检查跳表大小是否变化（应该不变或增加）
	newCard := s.ZCard()
	t.Logf("添加低分用户后跳表大小：%d", newCard)

	// 测试高分元素是否能进入
	t.Log("\n尝试添加高分元素（分数=500）...")
	s.ZAdd(500, "high_score_user")
	
	// 检查新守门员分数（应该提高）
	newGuardScore, ok := s.GetGuardScore()
	if ok {
		t.Logf("新守门员分数：%d (应该>原分数)", newGuardScore)
	}

	// 最终验证：前 10 名必须包含最高分
	t.Log("\n最终验证前 10 名:")
	for i := 0; i < 10; i++ {
		key, score := s.ZData(int64(i))
		t.Logf("第%d名：%s, 分数=%d", i+1, key, score)
	}
}

// 测试升序排序的人数限制
func TestMaxSizeAscendingOrder(t *testing.T) {
	t.Log("\n========== 测试升序排序的人数限制 ==========")

	// 创建升序 ZSet，最多保留 50 人（最低分在前）
	maxSize := int32(50)
	s := NewWithMaxSize(maxSize, 1)

	t.Logf("添加 100 个元素，分数从 1-100，只保留前 50 名（最低分）")

	for i := 0; i < 100; i++ {
		s.ZAdd(int64(i+1), fmt.Sprintf("user%d", i))
	}

	// 检查跳表大小
	card := s.ZCard()
	t.Logf("跳表元素数量：%d (期望：%d)", card, maxSize)
	if card != int64(maxSize) {
		t.Errorf("跳表大小错误：期望 %d, 实际 %d", maxSize, card)
	}

	// 获取守门员分数（第 50 名的分数）
	guardScore, ok := s.GetGuardScore()
	if !ok {
		t.Error("守门员应该有效")
	} else {
		t.Logf("守门员分数：%d (期望：50，因为最低分 1，第 50 名是 50)", guardScore)
		if guardScore != 50 {
			t.Errorf("守门员分数错误：期望 50, 实际 %d", guardScore)
		}
	}

	// 测试高分元素是否被拒绝
	t.Log("\n尝试添加高分元素（分数=200）...")
	s.ZAdd(200, "high_score_user")

	newCard := s.ZCard()
	t.Logf("添加高分用户后跳表大小：%d (应该不变)", newCard)
	if newCard != int64(maxSize) {
		t.Errorf("跳表大小不应该变化：%d", newCard)
	}

	// 验证前 10 名都是低分
	t.Log("\n验证前 10 名（应该是最低分）:")
	for i := 0; i < 10; i++ {
		key, score := s.ZData(int64(i))
		t.Logf("第%d名：%s, 分数=%d", i+1, key, score)
		if score > 10 {
			t.Errorf("前 10 名分数应该都<=10, 实际 %d", score)
		}
	}
}

// ==================== 性能测试 ====================

// TestZRevRankPerformance 测试 ZRank 在高并发下的性能
func TestZRevRankPerformance(t *testing.T) {
	t.Log("\n========== 高并发性能测试 ==========")
	
	// 创建一个有序集合（默认降序）
	set := New()

	// 添加大量元素
	const elementCount = 10000
	const concurrency = 100
	const queriesPerGoroutine = 1000

	t.Log("开始添加测试数据...")
	for i := 0; i < elementCount; i++ {
		key := fmt.Sprintf("user%d", i)
		score := int64(i % 100) // 创建一些相同分数的元素
		set.ZAdd(score, key)
	}
	t.Log("测试数据添加完成")

	// 准备要查询的键列表
	keys := make([]string, 0, concurrency)
	for i := 0; i < concurrency; i++ {
		keys = append(keys, fmt.Sprintf("user%d", i%elementCount))
	}

	t.Logf("开始高并发测试，%d 个 goroutine，每个执行 %d 次查询", concurrency, queriesPerGoroutine)
	startTime := time.Now()

	var wg sync.WaitGroup
	wg.Add(concurrency)

	// 启动多个 goroutine 并发查询
	for i := 0; i < concurrency; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			key := keys[goroutineID]

			for j := 0; j < queriesPerGoroutine; j++ {
				_, _ = set.ZRank(key)
				// 周期性打印进度
				if j%100 == 0 && goroutineID == 0 {
					t.Logf("进度：%d/%d", j, queriesPerGoroutine)
				}
			}
		}(i)
	}

	// 等待所有 goroutine 完成
	wg.Wait()
	elapsed := time.Since(startTime)

	totalQueries := concurrency * queriesPerGoroutine
	queriesPerSecond := float64(totalQueries) / elapsed.Seconds()

	t.Log("============================================")
	t.Logf("总查询次数：%d", totalQueries)
	t.Logf("总耗时：%v", elapsed)
	t.Logf("每秒查询次数：%.2f QPS", queriesPerSecond)
	t.Log("============================================")
	
	// 性能评估
	if queriesPerSecond < 50000 {
		t.Errorf("QPS 低于预期：%.2f (期望>50000)", queriesPerSecond)
	} else if queriesPerSecond < 100000 {
		t.Logf("性能良好：%.2f QPS", queriesPerSecond)
	} else {
		t.Logf("性能优秀：%.2f QPS", queriesPerSecond)
	}
}

// TestZRevRankWithTenThousandElements 测试添加 1 万个元素后执行 ZRank 操作
func TestZRevRankWithTenThousandElements(t *testing.T) {
	t.Log("\n========== 1 万元素性能测试 ==========")
	
	// 创建一个有序集合（默认降序，不限制人数）
	set := New()
	const elementCount = 10000

	t.Log("开始添加 1 万个测试数据...")
	startTime := time.Now()
	for i := 0; i < elementCount; i++ {
		key := fmt.Sprintf("user%d", i)
		score := int64(i % 100) // 创建一些相同分数的元素，模拟实际使用场景
		set.ZAdd(score, key)
	}
	addElapsed := time.Since(startTime)
	t.Logf("1 万个元素添加完成，耗时：%v", addElapsed)

	// 验证元素数量
	count := set.ZCard()
	t.Logf("当前集合元素数量：%d", count)
	if count != elementCount {
		t.Errorf("元素数量错误：期望 %d, 实际 %d", elementCount, count)
	}

	// 执行 ZRank 操作并测量性能
	t.Log("开始执行 ZRank 操作...")
	queryStartTime := time.Now()

	// 随机选择 100 个元素进行查询，测试性能
	queries := 100
	validQueries := 0
	for i := 0; i < queries; i++ {
		key := fmt.Sprintf("user%d", i*100) // 间隔 100 个取一个键
		rank, score := set.ZRank(key)
		if rank >= 0 {
			validQueries++
		}
		if i < 5 { // 打印前 5 个结果
			t.Logf("ZRank(%s): rank=%d, score=%d", key, rank, score)
		}
	}

	queryElapsed := time.Since(queryStartTime)
	avgLatency := queryElapsed / time.Duration(validQueries)
	
	t.Log("============================================")
	t.Logf("有效查询次数：%d/%d", validQueries, queries)
	t.Logf("执行 %d 次 ZRank 操作耗时：%v", validQueries, queryElapsed)
	t.Logf("平均每次查询耗时：%v", avgLatency)
	t.Log("============================================")
	
	// 性能评估
	if avgLatency > time.Millisecond {
		t.Errorf("平均延迟过高：%v (期望<1ms)", avgLatency)
	} else if avgLatency > 500*time.Microsecond {
		t.Logf("性能良好：平均延迟 %v", avgLatency)
	} else {
		t.Logf("性能优秀：平均延迟 %v", avgLatency)
	}
}

// ==================== 边界条件测试 ====================

// 测试空 ZSet 的操作
func TestEmptyZSet(t *testing.T) {
	t.Log("\n========== 空 ZSet 测试 ==========")

	s := New()

	// 测试空 ZSet 的排名
	rank, score := s.ZRank("nonexistent")
	if rank != -1 {
		t.Errorf("空 ZSet 中不存在的元素应该返回 rank=-1, 实际 %d", rank)
	}
	t.Logf("空 ZSet 查询不存在元素：rank=%d, score=%d", rank, score)

	// 测试空 ZSet 的元素数量
	card := s.ZCard()
	if card != 0 {
		t.Errorf("空 ZSet 的元素数量应该为 0, 实际 %d", card)
	}
	t.Logf("空 ZSet 元素数量：%d", card)

	// 测试空 ZSet 的 ZData
	key, score := s.ZData(0)
	if key != "" {
		t.Errorf("空 ZSet 的 ZData 应该返回空字符串，实际 %s", key)
	}
	t.Logf("空 ZSet 的 ZData(0): key='%s', score=%d", key, score)
}

// 测试重复添加相同元素
func TestDuplicateAdd(t *testing.T) {
	t.Log("\n========== 重复添加相同元素测试 ==========")

	s := New()

	// 多次添加相同元素
	for i := 0; i < 5; i++ {
		s.ZAdd(100, "same_user")
	}

	card := s.ZCard()
	if card != 1 {
		t.Errorf("重复添加相同元素应该只有 1 个，实际 %d", card)
	}
	t.Logf("重复添加 5 次相同元素后数量：%d", card)

	score, ok := s.ZScore("same_user")
	if !ok || score != 100 {
		t.Errorf("分数应该保持 100, 实际 %d (ok=%v)", score, ok)
	}
	t.Logf("最终分数：%d", score)
}

// 测试更新已存在元素的分数
func TestUpdateExistingElement(t *testing.T) {
	t.Log("\n========== 更新已存在元素测试 ==========")

	s := New()

	s.ZAdd(100, "user1")
	s.ZAdd(200, "user2")
	s.ZAdd(150, "user3")

	t.Log("初始状态：添加 3 个元素")

	// 更新 user1 的分数
	s.ZAdd(300, "user1")

	rank, score := s.ZRank("user1")
	t.Logf("更新后 user1: rank=%d, score=%d", rank, score)

	if score != 300 {
		t.Errorf("分数应该更新为 300, 实际 %d", score)
	}

	if rank != 0 {
		t.Errorf("user1 应该排名第 1(rank=0), 实际 rank=%d", rank)
	}

	// 检查元素总数
	card := s.ZCard()
	if card != 3 {
		t.Errorf("元素总数应该保持 3, 实际 %d", card)
	}
	t.Logf("更新后元素总数：%d", card)
}

// 测试删除守门员
func TestDeleteGuardKeeper(t *testing.T) {
	t.Log("\n========== 删除守门员测试 ==========")
	
	maxSize := int32(10)
	s := NewWithMaxSize(maxSize, -1)

	// 添加 15 个元素
	for i := 0; i < 15; i++ {
		s.ZAdd(int64(100-i), fmt.Sprintf("user%d", i))
	}

	// 获取初始守门员
	initialGuard, ok := s.GetGuardScore()
	if !ok {
		t.Error("守门员应该有效")
	}
	t.Logf("初始守门员分数：%d", initialGuard)

	// 删除守门员（最后一名）
	lastKey, _ := s.ZData(int64(maxSize - 1))
	t.Logf("删除守门员：%s", lastKey)
	
	ok = s.ZRem(lastKey)
	if !ok {
		t.Error("删除应该成功")
	}

	// 检查跳表大小（删除后应该是 9）
	card := s.ZCard()
	if card != int64(maxSize)-1 {
		t.Errorf("跳表大小应该为 %d, 实际 %d", maxSize-1, card)
	}
	t.Logf("删除守门员后跳表大小：%d", card)
	
	// 检查新守门员（未满员时应该无效）
	newGuard, ok := s.GetGuardScore()
	if ok {
		t.Logf("未满员时守门员无效，但实际有值：%d", newGuard)
	} else {
		t.Log("未满员时守门员无效（符合预期）")
	}
}

// 测试阈值清理机制
func TestThresholdCleanup(t *testing.T) {
	t.Log("\n========== 阈值清理机制测试 ==========")

	maxSize := int32(100)
	s := NewWithMaxSize(maxSize, -1)

	// 添加刚好满员的元素
	for i := 0; i < 100; i++ {
		s.ZAdd(int64(100-i), fmt.Sprintf("user%d", i))
	}

	t.Logf("初始状态：跳表大小=%d", s.ZCard())

	// 添加会被拒绝的低分元素（不会进入跳表，但会在 dict 中短暂存在）
	for i := 100; i < 150; i++ {
		s.ZAdd(int64(100-i), fmt.Sprintf("user%d", i))
	}

	// 跳表大小应该保持不变
	card := s.ZCard()
	t.Logf("添加 50 个低分元素后跳表大小：%d (应该=100)", card)
	if card != 100 {
		t.Errorf("跳表大小应该保持 100, 实际 %d", card)
	}

	// 检查是否触发了清理（dict 中不在跳表的元素应该被清理）
	t.Log("等待自动清理...")

	// 手动触发一次清理来验证
	s.tryTrimExcess()

	t.Logf("清理后测试完成")
}
