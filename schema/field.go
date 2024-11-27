package schema

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Field struct {
	Name              string
	Index             []int
	DBName            string
	Schema            *Schema //所在的父对象
	Embedded          *Schema //嵌入子对象
	FieldType         reflect.Type
	StructField       reflect.StructField
	IndirectFieldType reflect.Type
}

func (field *Field) JsonName() string {
	jsName := field.StructField.Tag.Get("json")
	if i := strings.Index(jsName, ","); i >= 0 {
		jsName = jsName[0:i]
	}
	if jsName == "" {
		return field.Name
	}
	return jsName
}

func (field *Field) GetEmbeddedFields() (r []*Field) {
	if !field.StructField.Anonymous || field.Embedded == nil {
		return
	}
	for _, ef := range field.Embedded.Fields {
		i := *ef
		v := &i
		if field.FieldType.Kind() == reflect.Struct {
			v.Index = append([]int{field.StructField.Index[0]}, v.Index...)
		} else {
			v.Index = append([]int{-field.StructField.Index[0] - 1}, v.Index...)
		}
		r = append(r, v)
	}
	return
}

func (field *Field) Get(value reflect.Value) reflect.Value {
	return field.ReflectValueOf(value)
}

func (field *Field) ReflectValueOf(value reflect.Value) reflect.Value {
	switch {
	case len(field.Index) == 1:
		return reflect.Indirect(value).Field(field.Index[0])
	case len(field.Index) == 2 && field.Index[0] >= 0 && field.FieldType.Kind() != reflect.Ptr:
		return reflect.Indirect(value).Field(field.Index[0]).Field(field.Index[1])
	default:
		v := reflect.Indirect(value)
		//logger.Debug("ReflectValueOf:%+v", v.Interface())
		for idx, fieldIdx := range field.Index {
			if fieldIdx >= 0 {
				v = v.Field(fieldIdx)
			} else {
				v = v.Field(-fieldIdx - 1)
			}
			//logger.Debug("ReflectValueOf:%+v", v.Interface())
			if v.Kind() == reflect.Ptr {
				if v.Type().Elem().Kind() == reflect.Struct {
					if v.IsNil() {
						v.Set(reflect.New(v.Type().Elem()))
					}
				}
				if idx < len(field.Index)-1 {
					v = v.Elem()
				}
			}
		}
		return v
	}
}

// Set valuer, setter when parse struct
func (field *Field) Set(value reflect.Value, v interface{}) error {
	switch field.FieldType.Kind() {
	case reflect.Bool:
		return field.SetBool(value, v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.SetInt(value, v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.SetUint(value, v)
	case reflect.Float32, reflect.Float64:
		return field.SetFloat(value, v)
	case reflect.String:
		return field.SetString(value, v)
	default:
		return field.fallbackSetter(value, v, field.Set)
	}
}

func (field *Field) fallbackSetter(value reflect.Value, v interface{}, setter func(reflect.Value, interface{}) error) (err error) {
	if v == nil {
		field.ReflectValueOf(value).Set(reflect.New(field.FieldType).Elem())
	} else {
		reflectV := reflect.ValueOf(v)
		// Optimal value type acquisition for v
		reflectValType := reflectV.Type()

		if reflectValType.AssignableTo(field.FieldType) {
			field.ReflectValueOf(value).Set(reflectV)
			return
		} else if reflectValType.ConvertibleTo(field.FieldType) {
			field.ReflectValueOf(value).Set(reflectV.Convert(field.FieldType))
			return
		} else if field.FieldType.Kind() == reflect.Ptr {
			fieldValue := field.ReflectValueOf(value)
			fieldType := field.FieldType.Elem()

			if reflectValType.AssignableTo(fieldType) {
				if !fieldValue.IsValid() {
					fieldValue = reflect.New(fieldType)
				} else if fieldValue.IsNil() {
					fieldValue.Set(reflect.New(fieldType))
				}
				fieldValue.Elem().Set(reflectV)
				return
			} else if reflectValType.ConvertibleTo(fieldType) {
				if fieldValue.IsNil() {
					fieldValue.Set(reflect.New(fieldType))
				}

				fieldValue.Elem().Set(reflectV.Convert(fieldType))
				return
			}
		}

		if reflectV.Kind() == reflect.Ptr {
			if reflectV.IsNil() {
				field.ReflectValueOf(value).Set(reflect.New(field.FieldType).Elem())
			} else {
				err = setter(value, reflectV.Elem().Interface())
			}
		} else if valuer, ok := v.(driver.Valuer); ok {
			if v, err = valuer.Value(); err == nil {
				err = setter(value, v)
			}
		} else {
			return fmt.Errorf("failed to set value %+v to field %s", v, field.Name)
		}
	}

	return
}

func (field *Field) SetBool(value reflect.Value, v interface{}) error {
	switch data := v.(type) {
	case bool:
		field.ReflectValueOf(value).SetBool(data)
	case *bool:
		if data != nil {
			field.ReflectValueOf(value).SetBool(*data)
		} else {
			field.ReflectValueOf(value).SetBool(false)
		}
	case int64:
		if data > 0 {
			field.ReflectValueOf(value).SetBool(true)
		} else {
			field.ReflectValueOf(value).SetBool(false)
		}
	case string:
		b, _ := strconv.ParseBool(data)
		field.ReflectValueOf(value).SetBool(b)
	default:
		return field.fallbackSetter(value, v, field.Set)
	}
	return nil
}
func (field *Field) SetInt(value reflect.Value, v interface{}) error {
	switch data := v.(type) {
	case int64:
		field.ReflectValueOf(value).SetInt(data)
	case int:
		field.ReflectValueOf(value).SetInt(int64(data))
	case int8:
		field.ReflectValueOf(value).SetInt(int64(data))
	case int16:
		field.ReflectValueOf(value).SetInt(int64(data))
	case int32:
		field.ReflectValueOf(value).SetInt(int64(data))
	case uint:
		field.ReflectValueOf(value).SetInt(int64(data))
	case uint8:
		field.ReflectValueOf(value).SetInt(int64(data))
	case uint16:
		field.ReflectValueOf(value).SetInt(int64(data))
	case uint32:
		field.ReflectValueOf(value).SetInt(int64(data))
	case uint64:
		field.ReflectValueOf(value).SetInt(int64(data))
	case float32:
		field.ReflectValueOf(value).SetInt(int64(data))
	case float64:
		field.ReflectValueOf(value).SetInt(int64(data))
	case []byte:
		return field.Set(value, string(data))
	case string:
		if data == "" {
			data = "0"
		}
		if i, err := strconv.ParseInt(data, 0, 64); err == nil {
			field.ReflectValueOf(value).SetInt(i)
		} else {
			return err
		}
	default:
		return field.fallbackSetter(value, v, field.Set)
	}
	return nil
}

func (field *Field) SetUint(value reflect.Value, v interface{}) error {
	switch data := v.(type) {
	case uint64:
		field.ReflectValueOf(value).SetUint(data)
	case uint:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case uint8:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case uint16:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case uint32:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case int64:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case int:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case int8:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case int16:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case int32:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case float32:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case float64:
		field.ReflectValueOf(value).SetUint(uint64(data))
	case []byte:
		return field.Set(value, string(data))
	case string:
		if data == "" {
			data = "0"
		}
		if i, err := strconv.ParseUint(data, 0, 64); err == nil {
			field.ReflectValueOf(value).SetUint(i)
		} else {
			return err
		}
	default:
		return field.fallbackSetter(value, v, field.Set)
	}
	return nil
}

func (field *Field) SetFloat(value reflect.Value, v interface{}) error {
	switch data := v.(type) {
	case float64:
		field.ReflectValueOf(value).SetFloat(data)
	case float32:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case int64:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case int:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case int8:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case int16:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case int32:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case uint:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case uint8:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case uint16:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case uint32:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case uint64:
		field.ReflectValueOf(value).SetFloat(float64(data))
	case []byte:
		return field.Set(value, string(data))
	case string:
		if data == "" {
			data = "0"
		}
		if i, err := strconv.ParseFloat(data, 64); err == nil {
			field.ReflectValueOf(value).SetFloat(i)
		} else {
			return err
		}
	default:
		return field.fallbackSetter(value, v, field.Set)
	}
	return nil
}

func (field *Field) SetString(value reflect.Value, v interface{}) error {
	switch data := v.(type) {
	case string:
		field.ReflectValueOf(value).SetString(data)
	case []byte:
		field.ReflectValueOf(value).SetString(string(data))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		field.ReflectValueOf(value).SetString(ToString(data))
	case float64, float32:
		field.ReflectValueOf(value).SetString(fmt.Sprintf("%v", data))
	default:
		return field.fallbackSetter(value, v, field.Set)
	}
	return nil
}
