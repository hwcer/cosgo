package utils

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/binary"
	"io"
)

func ZlibCompress(data []byte) []byte {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	_, _ = w.Write(data)
	_ = w.Close()
	return in.Bytes()
}

func ZlibUnCompress(data []byte) ([]byte, error) {
	b := bytes.NewReader(data)
	//🔴 NewReader 失败必须处理:否则 r 为 nil,io.ReadAll/Close 对 nil 调用直接 panic
	//(解压数据可能来自网络/存储,一包坏数据曾可打崩整个进程)
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = r.Close()
	}()
	undatas, err := io.ReadAll(r)
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
	//同 ZlibUnCompress:忽略 NewReader 错误会对 nil reader 调用方法直接 panic
	r, err := gzip.NewReader(b)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = r.Close()
	}()
	undatas, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return undatas, nil
}

// IntToBytes 整形转换成字节
func IntToBytes(n any) ([]byte, error) {
	var err error
	bytesBuffer := bytes.NewBuffer([]byte{})
	if v, ok := n.(int); ok {
		err = binary.Write(bytesBuffer, binary.BigEndian, int32(v))
	} else if v, ok := n.(float64); ok {
		err = binary.Write(bytesBuffer, binary.BigEndian, v)
	} else {
		err = binary.Write(bytesBuffer, binary.BigEndian, n)
	}
	if err != nil {
		return nil, err
	} else {
		return bytesBuffer.Bytes(), nil
	}

}

// IntToBuffer 将数字写入BUFFER, buffer := bytes.NewBuffer([]byte{})
func IntToBuffer(buffer *bytes.Buffer, n any) error {
	if v, ok := n.(int); ok {
		return binary.Write(buffer, binary.BigEndian, int32(v))
	} else if v, ok := n.(float64); ok {
		return binary.Write(buffer, binary.BigEndian, v)
	} else {
		return binary.Write(buffer, binary.BigEndian, n)
	}
}

// BytesToInt 字节转换成整形,n 必须是指针
// var a int32
// BytesToInt([]byte{1},&a)
func BytesToInt(b []byte, n any) error {
	bytesBuffer := bytes.NewBuffer(b)
	return binary.Read(bytesBuffer, binary.BigEndian, n)
}
