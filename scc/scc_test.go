package scc

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	CGO(test)
	CGO(testClose)
	_ = Wait(0)
}

func test(ctx context.Context) {
	t := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Ticker退出了\n")
			return
		case <-t.C:
			fmt.Printf("test Ticker\n")
		}
	}

}

func testClose(ctx context.Context) {
	time.AfterFunc(time.Second*5, func() {
		r := Cancel()
		fmt.Printf("倒计时结束，关闭程序:%v\n", r)
	})
}
