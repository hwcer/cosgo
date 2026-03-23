package zset

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestZAddPerformance 测试ZADD性能
func TestZAddPerformance(t *testing.T) {
	const concurrency = 10000
	const elementsPerGoroutine = 10
	const totalElements = concurrency * elementsPerGoroutine

	t.Logf("测试开始: %d 并发，每个goroutine添加 %d 个元素，共 %d 个元素", concurrency, elementsPerGoroutine, totalElements)

	// 测试ZSet
	t.Run("ZSet", func(t *testing.T) {
		set := New()
		start := time.Now()
		var wg sync.WaitGroup
		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < elementsPerGoroutine; j++ {
					key := fmt.Sprintf("user_%d_%d", goroutineID, j)
					score := int64(goroutineID*10000 + j)
					set.ZAdd(score, key)
				}
			}(i)
		}
		wg.Wait()
		totalTime := time.Since(start)
		t.Logf("ZSet: 总耗时 %v, 每秒操作数: %.2f", totalTime, float64(totalElements)/totalTime.Seconds())
	})
}

// TestZAddWithMaxSize 测试启用守门员机制的ZADD性能
func TestZAddWithMaxSize(t *testing.T) {
	const concurrency = 10000
	const maxSize = int32(1000)
	const testDuration = 1 * time.Second

	t.Logf("测试开始: %d 并发，MAXSIZE=%d，测试持续 %v", concurrency, maxSize, testDuration)

	// 测试ZSet
	t.Run("ZSet - 带守门员", func(t *testing.T) {
		set := NewWithMaxSize(maxSize)
		start := time.Now()
		endTime := start.Add(testDuration)
		var wg sync.WaitGroup
		wg.Add(concurrency)

		for i := 0; i < concurrency; i++ {
			go func(goroutineID int) {
				defer wg.Done()
				key := fmt.Sprintf("user_%d", goroutineID)
				score := int64(goroutineID)
				// 每秒ZADD一次
				for time.Now().Before(endTime) {
					score++
					set.ZAdd(score, key)
					time.Sleep(1 * time.Second)
					break // 只执行一次
				}
			}(i)
		}

		wg.Wait()
		totalTime := time.Since(start)
		finalSize := set.ZCard()
		t.Logf("ZSet(带守门员): 总耗时 %v, 最终大小: %d, 每秒操作数: %.2f",
			totalTime, finalSize, float64(concurrency)/totalTime.Seconds())
	})
}

// TestZAddWithIncreasingScore 测试倒序排列下，元素分数持续增加的性能
func TestZAddWithIncreasingScore(t *testing.T) {
	const totalElements = 10000
	const maxSize = int32(1000)

	t.Logf("测试开始: %d 个元素，倒序排列，MAXSIZE=%d，每个元素分数持续增加", totalElements, maxSize)

	// 测试ZSet
	t.Run("ZSet - 分数递增", func(t *testing.T) {
		// 创建倒序排列的ZSet（默认就是倒序）
		set := NewWithMaxSize(maxSize)

		// 初始化元素
		for i := 0; i < totalElements; i++ {
			key := fmt.Sprintf("user_%d", i)
			score := int64(i)
			set.ZAdd(score, key)
		}

		// 测试ZADD性能（分数持续增加）
		start := time.Now()
		var wg sync.WaitGroup
		wg.Add(totalElements)

		for i := 0; i < totalElements; i++ {
			go func(elementID int) {
				defer wg.Done()
				key := fmt.Sprintf("user_%d", elementID)
				score := int64(elementID + 10000) // 分数持续增加
				set.ZAdd(score, key)
			}(i)
		}

		wg.Wait()
		totalTime := time.Since(start)
		finalSize := set.ZCard()
		t.Logf("ZSet(分数递增): 总耗时 %v, 最终大小: %d, 每秒操作数: %.2f",
			totalTime, finalSize, float64(totalElements)/totalTime.Seconds())
	})
}
