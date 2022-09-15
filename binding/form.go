// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"fmt"
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/cosmo/schema"
	"io"
	"net/url"
	"reflect"
)

var formBindingSchema = schema.New()

func init() {
	j := &formBinding{}
	Register(MIMEPOSTForm, j)
}

type formBinding struct{}

func (f *formBinding) Name() string {
	return "form"
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
			field.Set(vf, data.GetString(field.DBName))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			field.Set(vf, data.GetInt64(field.DBName))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			field.Set(vf, data.GetInt64(field.DBName))
		case reflect.Float32, reflect.Float64:
			field.Set(vf, data.GetFloat64(field.DBName))
		}
	}
	return nil
}
func (f *formBinding) Bind(body io.Reader, i interface{}) (err error) {
	if body == nil {
		return fmt.Errorf("invalid request")
	}
	var b []byte
	b, err = io.ReadAll(body)
	if err != nil {
		return
	}
	return f.Unmarshal(b, i)
}
