package cosnet

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"cosgo/ioutil"
	"encoding/binary"
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

//整形转换成字节
func IntToBytes(n interface{}) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, n)
	return bytesBuffer.Bytes()
}

//字节转换成整形,n 必须是指针
// var a int32
// BytesToInt([]byte{1},&a)
func BytesToInt(b []byte, n interface{}) {
	bytesBuffer := bytes.NewBuffer(b)
	binary.Read(bytesBuffer, binary.BigEndian, n)
}

func ObjectIDPack(index, seed uint32) uint64 {
	return uint64(index)<<32 | uint64(seed)
}

//ObjectIDParse 返回ObjectIDPack中的index
func ObjectIDParse(id uint64) uint32 {
	return uint32(id >> 32)
}
