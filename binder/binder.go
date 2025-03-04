package binder

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"strings"
)

var binderMap = make(map[string]Binder)

//type Interface = Binder

type Binder interface {
	Name() string                        //JSON
	String() string                      // application/json
	Encode(io.Writer, interface{}) error //同Marshal
	Decode(io.Reader, interface{}) error //同Unmarshal
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
}

func New(t string) (b Binder) {
	return Get(t)
}

func Get(t string) (h Binder) {
	if st, ok := mimeTypes[strings.ToUpper(t)]; ok {
		return binderMap[st]
	}
	if st, ok := mimeNames[strings.ToLower(t)]; ok {
		return binderMap[st]
	}
	ct, _, err := mime.ParseMediaType(t)
	if err == nil {
		h = binderMap[ct]
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

func Encode(w io.Writer, i interface{}, t string) error {
	handle := Get(t)
	if handle == nil {
		return errors.New("name not exist")
	}
	return handle.Encode(w, i)
}

func Decode(r io.Reader, i interface{}, t string) error {
	handle := Get(t)
	if handle == nil {
		return errors.New("name not exist")
	}
	return handle.Decode(r, i)
}

func Marshal(i interface{}, t string) ([]byte, error) {
	handle := Get(t)
	if handle == nil {
		return nil, errors.New("type not exist")
	}
	return handle.Marshal(i)
}
func Unmarshal(b []byte, i interface{}, t string) error {
	handle := Get(t)
	if handle == nil {
		return errors.New("type not exist")
	}
	return handle.Unmarshal(b, i)
}
