package await

import (
	"sync"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	init := NewInitialize(func() error {
		time.Sleep(3 * time.Second)
		t.Logf("初始化完成")
		return nil
	})
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			t.Logf("goroutine start %d", n)
			_ = init.Verify()
			t.Logf("goroutine finish %d", n)
			wg.Done()
		}(i)
	}
	wg.Wait()
}