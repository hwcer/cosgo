package binder

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/schema"
	"io"
	"net/url"
	"reflect"
)

var formBindingSchema = schema.New()

var Form = &formBinding{}

func init() {
	_ = Register(MIMEPOSTForm, Form)
}

type formBinding struct{}

func (*formBinding) String() string {
	return MIMEPOSTForm
}

func (this *formBinding) Encode(w io.Writer, i interface{}) error {
	b, e := this.Marshal(i)
	if e != nil {
		return e
	}
	buf := bytes.NewBuffer(b)
	_, err := buf.WriteTo(w)
	return err
}

func (this *formBinding) Marshal(i interface{}) ([]byte, error) {
	vs := url.Values{}
	m, err := this.ToMap(i)
	if err != nil {
		return nil, err
	}
	for k, v := range m {
		vs.Set(k, v)
	}
	return []byte(vs.Encode()), nil
}

func (f *formBinding) Decode(r io.Reader, i interface{}) (err error) {
	var b []byte
	b, err = io.ReadAll(r)
	if err != nil {
		return
	}
	return f.Unmarshal(b, i)
}

func (f *formBinding) Unmarshal(b []byte, i interface{}) (err error) {
	if len(b) == 0 {
		return nil
	}
	var vs url.Values
	vs, err = url.ParseQuery(string(b))
	if err != nil {
		return
	}

	//values.Values
	if d, ok := i.(*values.Values); ok {
		for k, _ := range vs {
			d.Set(k, vs.Get(k))
		}
		return
	}
	//map[string]interface{}
	if d, ok := i.(*map[string]interface{}); ok {
		ss := *d
		for k, _ := range vs {
			ss[k] = vs.Get(k)
		}
		return
	}
	//struct
	data := values.Values{}
	for k, _ := range vs {
		data.Set(k, vs.Get(k))
	}
	vf := reflect.ValueOf(i)
	s, err := formBindingSchema.Parse(vf)
	if err != nil {
		return nil
	}
	for _, field := range s.Fields {
		switch field.IndirectFieldType.Kind() {
		case reflect.String:
			field.Set(vf, data.GetString(field.Name))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			field.Set(vf, data.GetInt64(field.Name))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			field.Set(vf, data.GetInt64(field.Name))
		case reflect.Float32, reflect.Float64:
			field.Set(vf, data.GetFloat64(field.Name))
		}
	}
	return nil
}
func (this *formBinding) ToMap(v any) (map[string]string, error) {
	vf := reflect.Indirect(reflect.ValueOf(v))
	switch vf.Kind() {
	case reflect.Struct:
		return this.fromStruct(v)
	case reflect.Map:
		return this.fromMap(vf)
	default:
		return nil, errors.New("unsupported type")
	}
}

func (this *formBinding) fromMap(vf reflect.Value) (map[string]string, error) {
	r := make(map[string]string)
	iter := vf.MapRange()
	for iter.Next() {
		k := fmt.Sprintf("%v", iter.Key().Interface())
		r[k] = fmt.Sprintf("%v", iter.Value().Interface())
	}
	return r, nil
}
func (this *formBinding) fromStruct(i any) (map[string]string, error) {
	vs := make(map[string]any)
	b, err := Json.Marshal(i)
	if err != nil {
		return nil, err
	}
	err = Json.Unmarshal(b, &vs)
	if err != nil {
		return nil, err
	}
	r := make(map[string]string)
	for k, _ := range vs {
		r[k] = fmt.Sprintf("%v", vs[k])
	}
	return r, nil
}
