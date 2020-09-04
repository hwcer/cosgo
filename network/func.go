package network

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
)

//启动服务器,根据
func Start(addr string, typ MsgType, handler IMsgHandler) (Server,error) {
	addrs := strings.Split(addr, "://")
	if addrs[0] == "tcp"  {
		return NewTcpServer(addrs[1],typ,handler)
	}else if addrs[0] == "udp" {
		//TODO UDP
	}else if addrs[0] == "ws" || addrs[0] == "wss" {
		//TODO wss
	}
	return nil,errors.New(fmt.Sprintf("server address error:%v",addr))
}


//关闭所有服务器
func Stop() {
	if !atomic.CompareAndSwapInt32(&stop, 0, 1) {
		Logger.Error("Server Stop error")
		return
	}
	wgp.Done()
	wgp.Wait()
	//所有服务器STOP
	Logger.Debug("ALL Server Stop")
}

func IsStop() bool {
	return stop == 1
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
				Logger.Error("%v", err)
			} else {
				handler[0](err)
			}
		}
	}()
	fun()
}

func Go(fun func(),unSafeGo ...bool)  {
	go func() {
		wgp.Add(1)
		defer wgp.Done()
		if len(unSafeGo) >0 && unSafeGo[0]{
			fun()
		}else {
			Try(fun)
		}
	}()
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
