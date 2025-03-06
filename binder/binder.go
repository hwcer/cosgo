package binder

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

var binderMap = make(map[string]Binder)

type Binder interface {
	Id() uint8                           // 1
	Name() string                        //JSON
	String() string                      //application/json
	Encode(io.Writer, interface{}) error //同Marshal
	Decode(io.Reader, interface{}) error //同Unmarshal
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
}

func New(t string) (b Binder) {
	return Get(t)
}

func ContentTypeFormat(c string) string {
	c = strings.ToLower(c)
	if i := strings.Index(c, ";"); i >= 0 {
		c = c[:i]
	}
	return strings.TrimSpace(c)
}

func Type(i any) (r *T) {
	switch v := i.(type) {
	case string:
		if strings.Contains(v, "/") {
			r = mimeTypes[ContentTypeFormat(v)]
		} else {
			r = mimeNames[strings.ToUpper(v)]
		}
	case int:
		r = mimeIds[uint8(v)]
	case uint8:
		r = mimeIds[v]
	default:
		r = nil
	}
	return
}

// Get 获取 string(name/type) uint8(id)
func Get(i any) (h Binder) {
	if t := Type(i); t != nil {
		h = binderMap[t.Type]
	}
	return
}

func Register(t string, handle Binder) error {
	if _, ok := binderMap[t]; ok {
		return fmt.Errorf("handle exist:%v", t)
	}
	binderMap[t] = handle
	return nil
}

func Encode(w io.Writer, i any, t any) error {
	handle := Get(t)
	if handle == nil {
		return errors.New("name not exist")
	}
	return handle.Encode(w, i)
}

func Decode(r io.Reader, i any, t any) error {
	handle := Get(t)
	if handle == nil {
		return errors.New("name not exist")
	}
	return handle.Decode(r, i)
}

func Marshal(i any, t any) ([]byte, error) {
	handle := Get(t)
	if handle == nil {
		return nil, errors.New("type not exist")
	}
	return handle.Marshal(i)
}
func Unmarshal(b []byte, i any, t any) error {
	handle := Get(t)
	if handle == nil {
		return errors.New("type not exist")
	}
	return handle.Unmarshal(b, i)
}
