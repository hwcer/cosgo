package zset

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestZRevRankPerformance 测试ZRevRank在高并发下的性能问题
func TestZRevRankPerformance(t *testing.T) {
	// 创建一个有序集合
	set := New(1) // 使用时间正序

	// 添加大量元素
	const elementCount = 10000
	const concurrency = 100
	const queriesPerGoroutine = 1000

	fmt.Println("开始添加测试数据...")
	for i := 0; i < elementCount; i++ {
		key := fmt.Sprintf("user%d", i)
		score := int64(i % 100) // 创建一些相同分数的元素
		set.ZAdd(score, key)
	}
	fmt.Println("测试数据添加完成")

	// 准备要查询的键列表
	keys := make([]string, 0, concurrency)
	for i := 0; i < concurrency; i++ {
		keys = append(keys, fmt.Sprintf("user%d", i%elementCount))
	}

	fmt.Printf("开始高并发测试，%d个goroutine，每个执行%d次查询\n", concurrency, queriesPerGoroutine)
	startTime := time.Now()

	var wg sync.WaitGroup
	wg.Add(concurrency)

	// 启动多个goroutine并发查询
	for i := 0; i < concurrency; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			key := keys[goroutineID]

			for j := 0; j < queriesPerGoroutine; j++ {
				_, _ = set.ZRevRank(key)
				// 周期性打印进度
				if j%100 == 0 && goroutineID == 0 {
					fmt.Printf("进度: %d/%d\n", j, queriesPerGoroutine)
				}
			}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()
	elapsed := time.Since(startTime)

	totalQueries := concurrency * queriesPerGoroutine
	queriesPerSecond := float64(totalQueries) / elapsed.Seconds()

	fmt.Printf("测试完成\n")
	fmt.Printf("总查询次数: %d\n", totalQueries)
	fmt.Printf("总耗时: %v\n", elapsed)
	fmt.Printf("每秒查询次数: %.2f QPS\n", queriesPerSecond)
}

// TestZRevRankWithTenThousandElements 测试添加1万个元素后执行ZRevRank操作
func TestZRevRankWithTenThousandElements(t *testing.T) {
	// 创建一个有序集合
	set := New(1) // 使用时间正序
	const elementCount = 10000

	fmt.Println("开始添加1万个测试数据...")
	startTime := time.Now()
	for i := 0; i < elementCount; i++ {
		key := fmt.Sprintf("user%d", i)
		score := int64(i % 100) // 创建一些相同分数的元素，模拟实际使用场景
		set.ZAdd(score, key)
	}
	addElapsed := time.Since(startTime)
	fmt.Printf("1万个元素添加完成，耗时: %v\n", addElapsed)

	// 验证元素数量
	count := set.ZCard()
	fmt.Printf("当前集合元素数量: %d\n", count)

	// 执行ZRevRank操作并测量性能
	fmt.Println("开始执行ZRevRank操作...")
	queryStartTime := time.Now()

	// 随机选择100个元素进行查询，测试性能
	queries := 100
	for i := 0; i < queries; i++ {
		key := fmt.Sprintf("user%d", i*100) // 间隔100个取一个键
		rank, score := set.ZRevRank(key)
		if i < 5 { // 打印前5个结果
			fmt.Printf("ZRevRank(%s): rank=%d, score=%d\n", key, rank, score)
		}
	}

	queryElapsed := time.Since(queryStartTime)
	fmt.Printf("执行%d次ZRevRank操作耗时: %v\n", queries, queryElapsed)
	fmt.Printf("平均每次查询耗时: %v\n", queryElapsed/time.Duration(queries))
}
