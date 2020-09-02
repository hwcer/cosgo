package network

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io/ioutil"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)
//关闭所有服务器
func Stop() {
	if !atomic.CompareAndSwapInt32(&stop, 0, 1) {
		Logger.Error("Server Stop error")
		return
	}
	//所有服务器STOP
	Logger.Debug("ALL Server Stop")
}
//FOR循环体中检查模块是否Runing
func loop() bool {
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

func Try(fun func(), handler ...func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			if len(handler) ==0 {
				Logger.Error("error catch:%v", err)
			} else {
				handler[0](err)
			}
		}
	}()
	fun()
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
