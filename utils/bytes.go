package utils

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/binary"
	"github.com/hwcer/cosgo/ioutil"
)

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

//IntToBytes 整形转换成字节
func IntToBytes(n interface{}) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	if v, ok := n.(int); ok {
		binary.Write(bytesBuffer, binary.BigEndian, int32(v))
	} else if v, ok := n.(float64); ok {
		binary.Write(bytesBuffer, binary.BigEndian, v)
	} else {
		binary.Write(bytesBuffer, binary.BigEndian, n)
	}
	return bytesBuffer.Bytes()
}

//字节转换成整形,n 必须是指针
// var a int32
// BytesToInt([]byte{1},&a)
func BytesToInt(b []byte, n interface{}) {
	bytesBuffer := bytes.NewBuffer(b)
	binary.Read(bytesBuffer, binary.BigEndian, n)
}
