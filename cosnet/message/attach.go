package message

import (
	"bytes"
	"errors"
	"github.com/hwcer/cosgo/utils"
)

var attachFieldsNum = 8            //最大字段长度
var attachFieldsName = [8]string{} //字段名
var attachValuesLength = [8]int{}  //字段值的长度

//SetAttachField length 占用字节
func SetAttachField(field string, length int) error {
	for i := 0; i < cap(attachFieldsName); i++ {
		if attachFieldsName[i] == "" {
			attachFieldsName[i] = field
			attachValuesLength[i] = length
			return nil
		}
	}
	return errors.New("Attach field num max")
}

type AttachIndex uint8

func (m *AttachIndex) Has(f int) bool {
	return (*m & (1 << f)) > 0
}

func (m *AttachIndex) Set(f int) {
	*m |= 1 << f
}
func (m *AttachIndex) Remove(f int) {
	if m.Has(f) {
		*m -= 1 << f
	}
}

//Size 返回attach长度
func (m *AttachIndex) Size() int {
	var len int
	for i := 0; i < attachFieldsNum; i++ {
		if m.Has(i) {
			len += attachValuesLength[i]
		}
	}
	return len
}

func NewAttach(index AttachIndex) *Attach {
	return &Attach{
		index:   index,
		dataset: [8][]byte{},
	}
}

type Attach struct {
	index   AttachIndex //8个byte
	dataset [8][]byte
}

func (a *Attach) IndexOf(key string) int {
	for i, v := range attachFieldsName {
		if v == key {
			return i
		}
	}
	return -1
}

func (a *Attach) Size() int {
	return a.index.Size()
}

// Parse 解析[]byte并填充字段
func (a *Attach) Parse(attach []byte) error {
	var (
		s int
		e int
	)
	for i := 0; i < attachFieldsNum; i++ {
		if a.index.Has(i) {
			e = s + attachValuesLength[i]
			if e > len(attach) {
				return errors.New("AttachIndex len error")
			}
			a.dataset[i] = attach[s:e]
			s = e
		}
	}
	return nil
}

//Bytes 生成二进制文件
func (a *Attach) Bytes() []byte {
	var b [][]byte
	for _, v := range a.dataset {
		if len(v) > 0 {
			b = append(b, v)
		}
	}
	return bytes.Join(b, []byte{})
}
func (a *Attach) Get(key string, val interface{}) bool {
	i := a.IndexOf(key)
	if i < 0 {
		return false
	}
	utils.BytesToInt(a.dataset[i], val)
	return true
}

func (a *Attach) Set(key string, val interface{}) error {
	i := a.IndexOf(key)
	if i < 0 {
		return errors.New("Attach key not exist")
	}
	b := utils.IntToBytes(val)
	if len(b) != attachValuesLength[i] {
		return errors.New("Attach val len error")
	}
	a.index.Set(i)
	a.dataset[i] = b
	return nil
}
func (a *Attach) Remove(key string) bool {
	i := a.IndexOf(key)
	if i < 0 {
		return false
	}
	a.index.Remove(i)
	a.dataset[i] = nil
	return true
}
