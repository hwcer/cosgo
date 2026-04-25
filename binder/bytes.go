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
	Binary: binary.BigEndian,
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
func (bb bytesBinding) Encode(w io.Writer, i any) error {
	data, err := bb.Marshal(i)
	if err != nil {
		return err
	}
	// 再写入数据
	_, err = w.Write(data)
	return err
}

func (bb bytesBinding) Decode(r io.Reader, i any) error {
	// 读取所有字节
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	return bb.Unmarshal(data, i)
}

func (bb bytesBinding) Marshal(i any) ([]byte, error) {
	// 如果已经是字节数组，直接返回
	if b, ok := i.([]byte); ok {
		return b, nil
	}

	v := reflect.ValueOf(i)
	if !v.IsValid() {
		return nil, fmt.Errorf("invalid data for marshal")
	}

	// 根据类型判断使用二进制还是 JSON 序列化
	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			return []byte{1}, nil
		}
		return []byte{0}, nil

	case reflect.Int8:
		buf := make([]byte, 1)
		buf[0] = byte(v.Int())
		return buf, nil

	case reflect.Int16:
		buf := make([]byte, 2)
		bb.Binary.PutUint16(buf, uint16(v.Int()))
		return buf, nil

	case reflect.Int32:
		buf := make([]byte, 4)
		bb.Binary.PutUint32(buf, uint32(v.Int()))
		return buf, nil

	case reflect.Int, reflect.Int64:
		buf := make([]byte, 8)
		bb.Binary.PutUint64(buf, uint64(v.Int()))
		return buf, nil

	case reflect.Uint8:
		buf := make([]byte, 1)
		buf[0] = byte(v.Uint())
		return buf, nil

	case reflect.Uint16:
		buf := make([]byte, 2)
		bb.Binary.PutUint16(buf, uint16(v.Uint()))
		return buf, nil

	case reflect.Uint32:
		buf := make([]byte, 4)
		bb.Binary.PutUint32(buf, uint32(v.Uint()))
		return buf, nil

	case reflect.Uint, reflect.Uint64:
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
		// 容器类型使用JSON
		return json.Marshal(i)
	}
}

func (bb bytesBinding) Unmarshal(b []byte, i any) error {
	if len(b) == 0 {
		return nil
	}

	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Pointer {
		return fmt.Errorf("invalid data for unmarshal: must be a pointer")
	}

	// 如果目标是字节数组，直接复制
	if dest, ok := i.(*[]byte); ok {
		*dest = make([]byte, len(b))
		copy(*dest, b)
		return nil
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

	case reflect.Int8:
		if len(b) < 1 {
			return fmt.Errorf("insufficient data for int8")
		}
		elem.SetInt(int64(int8(b[0])))
		return nil

	case reflect.Int16:
		if len(b) < 2 {
			return fmt.Errorf("insufficient data for int16")
		}
		// 先转 int16 保留符号位,再扩展到 int64;直接 int64(Uint16) 会把负数变成大正数
		elem.SetInt(int64(int16(bb.Binary.Uint16(b))))
		return nil

	case reflect.Int32:
		if len(b) < 4 {
			return fmt.Errorf("insufficient data for int32")
		}
		elem.SetInt(int64(int32(bb.Binary.Uint32(b))))
		return nil

	case reflect.Int, reflect.Int64:
		if len(b) < 8 {
			return fmt.Errorf("insufficient data for int64")
		}
		elem.SetInt(int64(bb.Binary.Uint64(b)))
		return nil

	case reflect.Uint8:
		if len(b) < 1 {
			return fmt.Errorf("insufficient data for uint8")
		}
		elem.SetUint(uint64(b[0]))
		return nil

	case reflect.Uint16:
		if len(b) < 2 {
			return fmt.Errorf("insufficient data for uint16")
		}
		elem.SetUint(uint64(bb.Binary.Uint16(b)))
		return nil

	case reflect.Uint32:
		if len(b) < 4 {
			return fmt.Errorf("insufficient data for uint32")
		}
		elem.SetUint(uint64(bb.Binary.Uint32(b)))
		return nil

	case reflect.Uint, reflect.Uint64:
		if len(b) < 8 {
			return fmt.Errorf("insufficient data for uint64")
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
		// 容器类型使用JSON
		return json.Unmarshal(b, i)
	}
}
