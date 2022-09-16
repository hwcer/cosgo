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
	w.Write(data)
	w.Close()
	return in.Bytes()
}

func ZlibUnCompress(data []byte) ([]byte, error) {
	b := bytes.NewReader(data)
	r, _ := zlib.NewReader(b)
	defer r.Close()
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
	r, _ := gzip.NewReader(b)
	defer r.Close()
	undatas, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return undatas, nil
}

//IntToBytes 整形转换成字节
func IntToBytes(n interface{}) ([]byte, error) {
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

//IntToBuffer 将数字写入BUFFER, buffer := bytes.NewBuffer([]byte{})
func IntToBuffer(buffer *bytes.Buffer, n interface{}) error {
	if v, ok := n.(int); ok {
		return binary.Write(buffer, binary.BigEndian, int32(v))
	} else if v, ok := n.(float64); ok {
		return binary.Write(buffer, binary.BigEndian, v)
	} else {
		return binary.Write(buffer, binary.BigEndian, n)
	}
}

//BytesToInt 字节转换成整形,n 必须是指针
// var a int32
// BytesToInt([]byte{1},&a)
func BytesToInt(b []byte, n interface{}) error {
	bytesBuffer := bytes.NewBuffer(b)
	return binary.Read(bytesBuffer, binary.BigEndian, n)
}
