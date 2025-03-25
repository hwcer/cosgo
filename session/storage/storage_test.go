package storage

import (
	"fmt"
	"github.com/hwcer/cosgo/random"
	"sync"
	"testing"
	"time"
)

var wg = &sync.WaitGroup{}
var data = New(10000)

func TestCacheCreate(t *testing.T) {
	wg.Add(1)
	go status()
	for i := 0; i < 5000; i++ {
		go agent()
	}
	wg.Wait()
}

func status() {
	wg.Add(1)
	defer wg.Done()
	for {
		fmt.Printf("=======================\nbucket:%d  ", len(data.bucket))
		fmt.Printf("size:%d ", data.Size())
		fmt.Printf("free:%d\n", data.Free())
		time.Sleep(1 * time.Second)
	}

}

func agent() {
	wg.Add(1)
	defer wg.Done()
	var tokens []string
	for {
		if n := len(tokens); n > 5 && random.Roll(5, int32(n)) > 10 {
			i := random.Roll(0, int32(n-1))
			if s := data.Delete(tokens[i]); s == nil {
				fmt.Printf("delete fail%s\n", tokens[i])
			} else {
				var t []string
				for k, v := range tokens {
					if k != int(i) {
						t = append(t, v)
					}
				}
				tokens = t
			}
		} else {
			if t := data.Create(nil); t != nil {
				tokens = append(tokens, t.Id())
			} else {
				fmt.Println("create fail")
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

}
