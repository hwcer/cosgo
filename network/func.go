package network

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
)

func Stop() {
	if !atomic.CompareAndSwapInt32(&stop, 0, 1) {
		Logger.Error("Server Stop error")
		return
	}

	close(stopChanForGo)
	for sc := 0; !waitAll.TryWait(); sc++ {
		Sleep(1)
		if sc >= Config.StopTimeout {
			Logger.Error("Server Stop Timeout")
			break
		}
	}
	Logger.Debug("Server Stop")
}

func IsStop() bool {
	return stop == 1
}

func IsRuning() bool {
	return stop == 0
}



func WaitForSystemExit() {
	var stopChanForSys = make(chan os.Signal, 1)
	signal.Notify(stopChanForSys, os.Interrupt, os.Kill, syscall.SIGTERM)
	select {
	case <-stopChanForSys:
		Stop()
	}
	close(stopChanForSys)
}

func Try(fun func(), handler func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			if handler == nil {
				Logger.Error("error catch:%v", err)
			} else {
				handler(err)
			}
		}
	}()
	fun()
}



func Go(fn func()) {
	pc := Config.PoolSize + 1
	select {
	case poolChan <- fn:
		return
	default:
		pc = atomic.AddInt32(&poolGoCount, 1)
		if pc > Config.PoolSize {
			atomic.AddInt32(&poolGoCount, -1)
		}
	}

	waitAll.Add(1)
	debugStr := simpleStack()
	id := atomic.AddUint32(&goid, 1)
	c := atomic.AddInt32(&gocount, 1)
	Logger.Debug("goroutine start id:%d count:%d from:%s", id, c, debugStr)
	go func() {
		Try(fn, nil)
		for pc <= Config.PoolSize {
			select {
			case <-stopChanForGo:
				pc = Config.PoolSize + 1
			case nfn := <-poolChan:
				Try(nfn, nil)
			}
		}
		waitAll.Done()
		c = atomic.AddInt32(&gocount, -1)
		Logger.Debug("goroutine end id:%d count:%d from:%s", id, c, debugStr)
	}()
}

func Go2(fn func(cstop chan struct{})) {
	Go(func() {
		Try(func() { fn(stopChanForGo) }, nil)
	})
}

func simpleStack() string {
	_, file, line, _ := runtime.Caller(2)
	i := strings.LastIndex(file, "/") + 1
	i = strings.LastIndex((string)(([]byte(file))[:i-1]), "/") + 1

	return fmt.Sprintf("%s:%d", (string)(([]byte(file))[i:]), line)
}


func ZlibCompress(data []byte) []byte {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	w.Write(data)
	w.Close()
	return in.Bytes()
}

func ZlibUnCompress(data []byte) ([]byte, error) {
	b := bytes.NewReader(data)
	r, _ := zlib.NewReader(b)
	defer r.Close()
	undatas, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return undatas, nil
}

func GZipCompress(data []byte) []byte {
	var in bytes.Buffer
	w := gzip.NewWriter(&in)
	w.Write(data)
	w.Close()
	return in.Bytes()
}

func GZipUnCompress(data []byte) ([]byte, error) {
	b := bytes.NewReader(data)
	r, _ := gzip.NewReader(b)
	defer r.Close()
	undatas, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return undatas, nil
}
