package zset

import (
	"sync"
)

// ZSetAsync 异步 ZSet 实现（基于通道）
type ZSetAsync struct {
	zset     *ZSet
	queue    chan func()
	wg       sync.WaitGroup
	stopChan chan struct{}
	started  bool
	mu       sync.RWMutex
}

// NewZSetAsync 创建异步 ZSet
func NewZSetAsync(queueSize int, order ...int8) *ZSetAsync {
	return &ZSetAsync{
		zset:     New(order...),
		queue:    make(chan func(), queueSize),
		stopChan: make(chan struct{}),
		started:  false,
	}
}

// NewZSetAsyncWithMaxSize 创建带人数限制的异步 ZSet
func NewZSetAsyncWithMaxSize(maxSize int32, queueSize int, order ...int8) *ZSetAsync {
	return &ZSetAsync{
		zset:     NewWithMaxSize(maxSize, order...),
		queue:    make(chan func(), queueSize),
		stopChan: make(chan struct{}),
		started:  false,
	}
}

// Start 启动异步处理器（必须调用）
func (z *ZSetAsync) Start(workerCount int) {
	z.mu.Lock()
	defer z.mu.Unlock()

	if z.started {
		return
	}

	z.started = true

	// 强制使用单 worker，确保跳表操作的串行化
	for i := 0; i < 1; i++ {
		z.wg.Add(1)
		go z.worker()
	}
}

// Stop 停止异步处理器
func (z *ZSetAsync) Stop() {
	close(z.stopChan)
	z.wg.Wait()
}

// worker 工作协程
func (z *ZSetAsync) worker() {
	defer z.wg.Done()

	for {
		select {
		case task := <-z.queue:
			if task != nil {
				task()
			}
		case <-z.stopChan:
			return
		}
	}
}

// ZAdd 异步添加元素（非阻塞）
func (z *ZSetAsync) ZAdd(score int64, key string) *future {
	f := &future{
		done: make(chan struct{}),
	}

	z.queue <- func() {
		defer close(f.done)
		z.zset.ZAdd(score, key)
	}

	return f
}

// ZIncr 异步增加分数
func (z *ZSetAsync) ZIncr(score int64, key string) *futureInt64 {
	f := &futureInt64{
		future: future{
			done: make(chan struct{}),
		},
	}

	z.queue <- func() {
		defer close(f.done)
		f.value = z.zset.ZIncr(score, key)
	}

	return f
}

// ZRem 异步删除元素
func (z *ZSetAsync) ZRem(key string) *futureBool {
	f := &futureBool{
		future: future{
			done: make(chan struct{}),
		},
	}

	z.queue <- func() {
		defer close(f.done)
		f.value = z.zset.ZRem(key)
	}

	return f
}

// ZRank 异步获取排名
func (z *ZSetAsync) ZRank(key string) *futureRank {
	f := &futureRank{
		future: future{
			done: make(chan struct{}),
		},
	}

	z.queue <- func() {
		defer close(f.done)
		f.rank, f.score = z.zset.ZRank(key)
	}

	return f
}

// ZScore 异步获取分数
func (z *ZSetAsync) ZScore(key string) *futureScore {
	f := &futureScore{
		future: future{
			done: make(chan struct{}),
		},
	}

	z.queue <- func() {
		defer close(f.done)
		f.score, f.ok = z.zset.ZScore(key)
	}

	return f
}

// ZData 异步获取指定排名的元素
func (z *ZSetAsync) ZData(rank int64) *futureData {
	f := &futureData{
		future: future{
			done: make(chan struct{}),
		},
	}

	z.queue <- func() {
		defer close(f.done)
		f.key, f.score = z.zset.ZData(rank)
	}

	return f
}

// ZRange 异步遍历范围（回调方式）
func (z *ZSetAsync) ZRange(start, end int64, callback func(int64, string)) *future {
	f := &future{
		done: make(chan struct{}),
	}

	z.queue <- func() {
		defer close(f.done)
		z.zset.ZRange(start, end, callback)
	}

	return f
}

// ZCard 异步获取元素数量
func (z *ZSetAsync) ZCard() *futureInt64 {
	f := &futureInt64{
		future: future{
			done: make(chan struct{}),
		},
	}

	z.queue <- func() {
		defer close(f.done)
		f.value = z.zset.ZCard()
	}

	return f
}

// GetGuardScore 异步获取守门员分数
func (z *ZSetAsync) GetGuardScore() *futureScore {
	f := &futureScore{
		future: future{
			done: make(chan struct{}),
		},
	}

	z.queue <- func() {
		defer close(f.done)
		f.score, f.ok = z.zset.GetGuardScore()
	}

	return f
}

// SetMaxSize 异步设置最大人数
func (z *ZSetAsync) SetMaxSize(maxSize int32) *future {
	f := &future{
		done: make(chan struct{}),
	}

	z.queue <- func() {
		defer close(f.done)
		z.zset.SetMaxSize(maxSize)
	}

	return f
}
