package session

import (
	"fmt"
	"github.com/hwcer/cosgo/values"
	"strconv"
	"sync"
	"testing"
	"time"
)

var store *Memory
var wait = sync.WaitGroup{}

func init() {
	store = NewMemory()
	_ = store.Start()
}

func TestNewMemory(t *testing.T) {
	wait.Add(1)
	for i := 1; i <= 5000; i++ {
		go worker(i)
	}
	wait.Done()
	wait.Wait()
}

func worker(i int) {
	wait.Add(1)
	defer wait.Done()
	uuid := strconv.Itoa(i)
	vals := values.Values{}
	vals["uid"] = uuid
	mid, err := store.Create(uuid, vals, 3600, false)
	if err != nil {
		fmt.Printf("Create error:%v\n", err)
		return
	}
	for {
		_, _, e := store.Get(mid, false)
		if e != nil {
			fmt.Printf("Get error:%v\n", e)
			return
		}
		ch := time.After(time.Millisecond)
		<-ch
	}

}
