package schema

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type concurrentModelA struct {
	ID   int64  `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"`
}

type concurrentModelB struct {
	ID    int64   `json:"id" bson:"_id"`
	Value float64 `json:"value" bson:"value"`
}

// TestConcurrentParse_Waiter 模拟高并发首次解析: 一组 goroutine 同时 Parse 同一类型,
// 其中一个"胜出"真正构建,其余应在 schema 构建完成后立即被 chan 唤醒,
// 而不是以 1ms 轮询等待。
func TestConcurrentParse_Waiter(t *testing.T) {
	opts := New()
	const workers = 50
	var wg sync.WaitGroup
	var ok int32
	start := make(chan struct{})

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			s, err := opts.Parse(&concurrentModelA{})
			if err == nil && s != nil && len(s.Fields) > 0 {
				atomic.AddInt32(&ok, 1)
			}
		}()
	}
	t0 := time.Now()
	close(start)
	wg.Wait()
	dur := time.Since(t0)

	if got := atomic.LoadInt32(&ok); got != workers {
		t.Fatalf("only %d/%d goroutines got a valid schema", got, workers)
	}
	// 宽松上限: 50 个并发的首次解析应当远低于 100ms(chan 唤醒是 μs 级)
	if dur > 100*time.Millisecond {
		t.Errorf("concurrent first-parse took %v (expected << 100ms via chan signal)", dur)
	}
}

// TestWarm_NoWaitOnHot 验证 Warm 后后续请求不再触发等待: initDone 已 close,
// waitSchemaInit 快路径立即返回。
func TestWarm_NoWaitOnHot(t *testing.T) {
	opts := New()
	if err := WarmWithOptions(opts, &concurrentModelA{}, &concurrentModelB{}); err != nil {
		t.Fatal(err)
	}
	// 后续获取应无等待
	t0 := time.Now()
	for i := 0; i < 10000; i++ {
		_, _ = opts.Parse(&concurrentModelA{})
		_, _ = opts.Parse(&concurrentModelB{})
	}
	dur := time.Since(t0)
	// 20000 次纯缓存命中应远低于 100ms
	if dur > 100*time.Millisecond {
		t.Errorf("20000 cached Parse calls took %v (expected << 100ms)", dur)
	}
}
