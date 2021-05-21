package message

import (
	"bytes"
	"errors"
	"github.com/hwcer/cosgo/utils"
)

type AttachIndex uint16

var AttachFields = [16]string{} //字段名
var AttachValues = [16]int{}    //字段值的长度

var attachFieldsNum int = 16

//size 字段字节数
func SetAttachField(field string, size int) error {
	for i := 0; i < cap(AttachFields); i++ {
		if AttachFields[i] == "" {
			AttachFields[i] = field
			AttachValues[i] = size
			return nil
		}
	}
	return errors.New("Attach field num max")
}

func (m *AttachIndex) Has(f int) bool {
	return (*m & 1 << f) > 0
}

func (m *AttachIndex) Add(f int) {
	*m |= 1 << f
}
func (m *AttachIndex) Del(f int) {
	if m.Has(f) {
		*m -= 1 << f
	}
}

func NewAttach() *Attach {
	return &Attach{
		dataset: [16]int32{},
	}
}

type Attach struct {
	index   AttachIndex //8个byte
	dataset [16]int32
}

func (m *Attach) IndexOf(key string) int {
	for i, v := range AttachFields {
		if v == key {
			return i
		}
	}
	return -1
}

//parse 解析[]byte并填充字段
func (m *Attach) Parse(attach []byte) error {
	if len(attach) < 8 {
		return errors.New("attach len error")
	}
	utils.BytesToInt(attach[0:8], &m.index)
	var s int = 8
	var e int = 8
	for i := 0; i < 8; i++ {
		if m.index.Has(i) {
			e += 32
			if e > len(attach) {
				return errors.New("attach len error")
			}
			utils.BytesToInt(attach[s:e], &m.dataset[i])
			s = e
		}
	}
	return nil
}

//Bytes 生成二进制文件
func (r *Attach) Bytes() []byte {
	var b [][]byte
	b = append(b, utils.IntToBytes(r.index))
	for i, v := range r.dataset {
		if v != 0 {
			b = append(b, utils.IntToBytes(v))
			r.index.Add(i)
		}
	}
	return bytes.Join(b, []byte{})
}
func (r *Attach) Get(key string) (int32, bool) {
	i := r.IndexOf(key)
	if i < 0 {
		return 0, false
	}
	return r.dataset[i], true
}

func (r *Attach) Set(key string, val int32) bool {
	i := r.IndexOf(key)
	if i < 0 {
		return false
	}
	r.index.Add(i)
	r.dataset[i] = val
	return true
}
