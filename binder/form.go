package binder

import (
	"bytes"
	"errors"
	"fmt"
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
	return f.UnmarshalFromValues(vs, i)
}

func (f *formBinding) UnmarshalFromValues(vs url.Values, i any) (err error) {
	if len(vs) == 0 {
		return nil
	}
	vf := reflect.ValueOf(i)
	if iv := reflect.Indirect(vf); iv.Kind() == reflect.Map {
		for k, _ := range vs {
			if v := vs.Get(k); v != "" {
				iv.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
			}
		}
		return nil
	}

	s, err := formBindingSchema.Parse(vf)
	if err != nil {
		return nil
	}
	for _, field := range s.Fields {
		k := field.JsonName()
		if v := vs.Get(k); v != "" {
			if err = field.Set(vf, v); err != nil {
				return err
			}
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
