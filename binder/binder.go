package binder

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type EncodingType uint8

const (
	EncodingTypeXml EncodingType = iota + 1
	EncodingTypeJson
	EncodingTypeYaml
	EncodingTypeProtoBuf
	EncodingTypeUrlEncoded
)

var binderMap = make(map[EncodingType]Interface)

type Interface interface {
	String() string
	Encode(io.Writer, interface{}) error //同Marshal
	Decode(io.Reader, interface{}) error //同Unmarshal
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
}

func New(t EncodingType) (b *Binder) {
	h := Handle(t)
	if h != nil {
		b = &Binder{handle: h}
	}
	return
}

func Handle(t EncodingType) (h Interface) {
	return binderMap[t]
}

func Register(t EncodingType, handle Interface) error {
	if _, ok := binderMap[t]; ok {
		return fmt.Errorf("handle exist:%v", t)
	}
	binderMap[t] = handle
	return nil
}

func Encode(w io.Writer, i interface{}, t EncodingType) error {
	handle := Handle(t)
	if handle == nil {
		return errors.New("name not exist")
	}
	return handle.Encode(w, i)
}

func Decode(r io.Reader, i interface{}, t EncodingType) error {
	handle := Handle(t)
	if handle == nil {
		return errors.New("name not exist")
	}
	return handle.Decode(r, i)
}

func Marshal(i interface{}, t EncodingType) ([]byte, error) {
	handle := Handle(t)
	if handle == nil {
		return nil, errors.New("type not exist")
	}
	return handle.Marshal(i)
}
func Unmarshal(b []byte, i interface{}, t EncodingType) error {
	handle := Handle(t)
	if handle == nil {
		return errors.New("type not exist")
	}
	return handle.Unmarshal(b, i)
}

type Binder struct {
	handle Interface
}

func (*Binder) String() string {
	return "Binder"
}

func (this *Binder) Encode(w io.Writer, i interface{}) (err error) {
	switch v := i.(type) {
	case []byte:
		_, err = w.Write(v)
	case string:
		_, err = w.Write([]byte(v))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		err = binary.Write(w, binary.BigEndian, v)
	default:
		err = this.handle.Encode(w, i)
	}
	return
}

func (this *Binder) Decode(r io.Reader, i interface{}) (err error) {
	switch v := i.(type) {
	case []byte:
		_, err = io.ReadFull(r, v)
	case *string:
		var b []byte
		if b, err = io.ReadAll(r); err == nil {
			*v = string(b)
		}
	case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, *float32, *float64, *bool, *complex64, *complex128:
		err = binary.Read(r, binary.BigEndian, i)
	default:
		err = this.handle.Decode(r, i)
	}
	return
}

// Marshal 将一个对象放入Message.data
func (this *Binder) Marshal(i interface{}) (b []byte, err error) {
	switch v := i.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		buf := bytes.NewBuffer(b)
		if err = binary.Write(buf, binary.BigEndian, v); err == nil {
			b = buf.Bytes()
		}
	default:
		b, err = this.handle.Marshal(i)
	}
	return
}

// Unmarshal 解析Message body
func (this *Binder) Unmarshal(b []byte, i interface{}) (err error) {
	switch v := i.(type) {
	case *string:
		*v = string(b)
	case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, *float32, *float64, *bool, *complex64, *complex128:
		err = binary.Read(bytes.NewReader(b), binary.BigEndian, i)
	default:
		err = this.handle.Unmarshal(b, i)
	}
	return
}
