package binder

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"
)

var Bytes = bytesBinding{
	Binary: binary.LittleEndian,
}

func init() {
	_ = Register(MIMEBytes, Bytes)
}

type bytesBinding struct {
	Binary binary.ByteOrder
}

func (bytesBinding) Id() uint8 {
	return Type(MIMEBytes).Id
}

func (bytesBinding) Name() string {
	return Type(MIMEBytes).Name
}
func (bytesBinding) String() string {
	return MIMEBytes
}
func (bb bytesBinding) Encode(w io.Writer, i interface{}) error {
	data, err := bb.Marshal(i)
	if err != nil {
		return err
	}
	// 再写入数据
	_, err = w.Write(data)
	return err
}

func (bb bytesBinding) Decode(r io.Reader, i interface{}) error {
	// 读取所有字节
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	return bb.Unmarshal(data, i)
}

func (bb bytesBinding) Marshal(i interface{}) ([]byte, error) {
	v := reflect.ValueOf(i)
	if !v.IsValid() {
		return json.Marshal(i)
	}

	// 根据类型判断使用二进制还是 JSON 序列化
	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			return []byte{1}, nil
		}
		return []byte{0}, nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		buf := make([]byte, 8)
		bb.Binary.PutUint64(buf, uint64(v.Int()))
		return buf, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		buf := make([]byte, 8)
		bb.Binary.PutUint64(buf, v.Uint())
		return buf, nil

	case reflect.Float32:
		buf := make([]byte, 4)
		bb.Binary.PutUint32(buf, math.Float32bits(float32(v.Float())))
		return buf, nil

	case reflect.Float64:
		buf := make([]byte, 8)
		bb.Binary.PutUint64(buf, math.Float64bits(v.Float()))
		return buf, nil

	case reflect.String:
		return []byte(v.String()), nil

	default:
		// 容器类型使用 JSON
		return json.Marshal(i)
	}
}

func (bb bytesBinding) Unmarshal(b []byte, i interface{}) error {
	if len(b) == 0 {
		return nil
	}

	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return json.Unmarshal(b, i)
	}

	elem := v.Elem()

	// 根据类型判断使用二进制还是 JSON 反序列化
	switch elem.Kind() {
	case reflect.Bool:
		if len(b) > 0 && b[0] != 0 {
			elem.SetBool(true)
		} else {
			elem.SetBool(false)
		}
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if len(b) < 8 {
			return fmt.Errorf("insufficient data for int")
		}
		elem.SetInt(int64(bb.Binary.Uint64(b)))
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if len(b) < 8 {
			return fmt.Errorf("insufficient data for uint")
		}
		elem.SetUint(bb.Binary.Uint64(b))
		return nil

	case reflect.Float32:
		if len(b) < 4 {
			return fmt.Errorf("insufficient data for float32")
		}
		elem.SetFloat(float64(math.Float32frombits(bb.Binary.Uint32(b))))
		return nil

	case reflect.Float64:
		if len(b) < 8 {
			return fmt.Errorf("insufficient data for float64")
		}
		elem.SetFloat(math.Float64frombits(bb.Binary.Uint64(b)))
		return nil

	case reflect.String:
		elem.SetString(string(b))
		return nil

	default:
		// 容器类型使用 JSON
		return json.Unmarshal(b, i)
	}
}
