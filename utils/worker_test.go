package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestWorker(t *testing.T) {
	w := NewWorker(2, workerHandle)
	for i := 0; i < 5; i++ {
		w.Emit(i)
	}
	<-time.After(time.Second * 5)
	w.Close()
}

func workerHandle(m interface{}) {
	fmt.Printf("workerHandle:%v\n", m)
}
