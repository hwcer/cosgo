package cosnet

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

//启动服务器,根据
func NewServer(addr string, typ MsgType, handler IMsgHandler) (Server, error) {
	addrs := strings.Split(addr, "://")
	if addrs[0] == "tcp" {
		return NewTcpServer(addrs[1], typ, handler)
	} else if addrs[0] == "udp" {
		//TODO UDP
	} else if addrs[0] == "ws" || addrs[0] == "wss" {
		//TODO wss
	}
	return nil, errors.New(fmt.Sprintf("server address error:%v", addr))
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
