package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

var mux sync.Mutex
var wg sync.WaitGroup

func TestSocket(t *testing.T) {
	mainch := make(chan int)
	go mlock(mainch)
	for i := 1; i <= 100; i++ {
		wg.Add(1)
		go clock(i)
		time.Sleep(time.Millisecond)
	}
	mainch <- 1
	wg.Wait()
}

func mlock(ch chan int) {
	mux.Lock()
	defer mux.Unlock()
	<-ch
	fmt.Printf("主协程释放锁\n")
}

func clock(i int) {
	mux.Lock()
	defer wg.Done()
	defer mux.Unlock()

	fmt.Printf("子协程获得锁%v\n", i)
}
