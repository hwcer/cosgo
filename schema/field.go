package schema

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
)

type Field struct {
	Name              string
	Index             []int
	DBName            string
	FieldType         reflect.Type
	IndirectFieldType reflect.Type
	StructField       reflect.StructField
	Schema            *Schema //所在的父对象
	EmbeddedSchema    *Schema //嵌入子对象

	//OwnerSchema    *Schema
	ReflectValueOf func(reflect.Value) reflect.Value
	//ValueOf        func(reflect.Value) (value interface{}, zero bool)
	Set func(reflect.Value, interface{}) error
}

func (field *Field) GetFields() (r []*Field) {
	if field.StructField.Anonymous {
		for _, ef := range field.EmbeddedSchema.Fields {
			i := *ef
			v := &i
			if field.FieldType.Kind() == reflect.Struct {
				v.Index = append([]int{field.StructField.Index[0]}, v.StructField.Index...)
			} else {
				v.Index = append([]int{-field.StructField.Index[0] - 1}, v.StructField.Index...)
			}
			r = append(r, v)
		}
	} else {
		r = append(r, field)
	}
	return
}

func (schema *Schema) ParseField(fieldStruct reflect.StructField) *Field {
	//var err error

	field := &Field{
		Name:              fieldStruct.Name,
		FieldType:         fieldStruct.Type,
		IndirectFieldType: fieldStruct.Type,
		StructField:       fieldStruct,
		Schema:            schema,
	}

	for field.IndirectFieldType.Kind() == reflect.Ptr {
		field.IndirectFieldType = field.IndirectFieldType.Elem()
	}

	fieldValue := reflect.New(field.IndirectFieldType)
	if dbName := fieldStruct.Tag.Get("bson"); dbName != "" && dbName != "inline" {
		field.DBName = dbName
	}

	if fieldStruct.Anonymous {
		kind := reflect.Indirect(fieldValue).Kind()
		switch kind {
		case reflect.Struct:
			var err error
			if field.EmbeddedSchema, err = getOrParse(fieldValue.Interface(), schema.options); err != nil {
				schema.err = err
			}
		case reflect.Invalid, reflect.Uintptr, reflect.Array, reflect.Chan, reflect.Func, reflect.Interface,
			reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer, reflect.Complex64, reflect.Complex128:
			schema.err = fmt.Errorf("invalid embedded struct for %s's field %s, should be struct, but got %v", field.Schema.Name, field.Name, field.FieldType)
		}
	} else {
		field.Index = field.StructField.Index
	}

	return field
}

// create valuer, setter when parse struct
func (field *Field) setupValuerAndSetter() {
	// ValueOf
	//switch {
	//case len(field.StructField.Index) == 1:
	//	field.ValueOf = func(value reflect.Value) (interface{}, bool) {
	//		fieldValue := reflect.Indirect(value).Field(field.StructField.Index[0])
	//		return fieldValue.Interface(), fieldValue.IsZero()
	//	}
	//case len(field.StructField.Index) == 2 && field.StructField.Index[0] >= 0:
	//	field.ValueOf = func(value reflect.Value) (interface{}, bool) {
	//		fieldValue := reflect.Indirect(value).Field(field.StructField.Index[0]).Field(field.StructField.Index[1])
	//		return fieldValue.Interface(), fieldValue.IsZero()
	//	}
	//default:
	//	field.ValueOf = func(value reflect.Value) (interface{}, bool) {
	//		v := reflect.Indirect(value)
	//
	//		for _, idx := range field.StructField.Index {
	//			if idx >= 0 {
	//				v = v.Field(idx)
	//			} else {
	//				v = v.Field(-idx - 1)
	//
	//				if v.Type().Elem().Kind() != reflect.Struct {
	//					return nil, true
	//				}
	//
	//				if !v.IsNil() {
	//					v = v.Elem()
	//				} else {
	//					return nil, true
	//				}
	//			}
	//		}
	//		return v.Interface(), v.IsZero()
	//	}
	//}

	// ReflectValueOf
	switch {
	case len(field.Index) == 1:
		field.ReflectValueOf = func(value reflect.Value) reflect.Value {
			return reflect.Indirect(value).Field(field.Index[0])
		}
	case len(field.Index) == 2 && field.Index[0] >= 0 && field.FieldType.Kind() != reflect.Ptr:
		field.ReflectValueOf = func(value reflect.Value) reflect.Value {
			return reflect.Indirect(value).Field(field.Index[0]).Field(field.Index[1])
		}
	default:
		field.ReflectValueOf = func(value reflect.Value) reflect.Value {
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

	fallbackSetter := func(value reflect.Value, v interface{}, setter func(reflect.Value, interface{}) error) (err error) {
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

	// Set
	switch field.FieldType.Kind() {
	case reflect.Bool:
		field.Set = func(value reflect.Value, v interface{}) error {
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
				return fallbackSetter(value, v, field.Set)
			}
			return nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.Set = func(value reflect.Value, v interface{}) (err error) {
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
				if i, err := strconv.ParseInt(data, 0, 64); err == nil {
					field.ReflectValueOf(value).SetInt(i)
				} else {
					return err
				}
			//case time.Time:
			//	if field.AutoCreateTime == UnixNanosecond || field.AutoUpdateTime == UnixNanosecond {
			//		field.ReflectValueOf(value).SetInt(data.UnixNano())
			//	} else if field.AutoCreateTime == UnixMillisecond || field.AutoUpdateTime == UnixMillisecond {
			//		field.ReflectValueOf(value).SetInt(data.UnixNano() / 1e6)
			//	} else {
			//		field.ReflectValueOf(value).SetInt(data.Unix())
			//	}
			//case *time.Time:
			//	if data != nil {
			//		if field.AutoCreateTime == UnixNanosecond || field.AutoUpdateTime == UnixNanosecond {
			//			field.ReflectValueOf(value).SetInt(data.UnixNano())
			//		} else if field.AutoCreateTime == UnixMillisecond || field.AutoUpdateTime == UnixMillisecond {
			//			field.ReflectValueOf(value).SetInt(data.UnixNano() / 1e6)
			//		} else {
			//			field.ReflectValueOf(value).SetInt(data.Unix())
			//		}
			//	} else {
			//		field.ReflectValueOf(value).SetInt(0)
			//	}
			default:
				return fallbackSetter(value, v, field.Set)
			}
			return err
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field.Set = func(value reflect.Value, v interface{}) (err error) {
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
			//case time.Time:
			//	if field.AutoCreateTime == UnixNanosecond || field.AutoUpdateTime == UnixNanosecond {
			//		field.ReflectValueOf(value).SetUint(uint64(data.UnixNano()))
			//	} else if field.AutoCreateTime == UnixMillisecond || field.AutoUpdateTime == UnixMillisecond {
			//		field.ReflectValueOf(value).SetUint(uint64(data.UnixNano() / 1e6))
			//	} else {
			//		field.ReflectValueOf(value).SetUint(uint64(data.Unix()))
			//	}
			case string:
				if i, err := strconv.ParseUint(data, 0, 64); err == nil {
					field.ReflectValueOf(value).SetUint(i)
				} else {
					return err
				}
			default:
				return fallbackSetter(value, v, field.Set)
			}
			return err
		}
	case reflect.Float32, reflect.Float64:
		field.Set = func(value reflect.Value, v interface{}) (err error) {
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
				if i, err := strconv.ParseFloat(data, 64); err == nil {
					field.ReflectValueOf(value).SetFloat(i)
				} else {
					return err
				}
			default:
				return fallbackSetter(value, v, field.Set)
			}
			return err
		}
	case reflect.String:
		field.Set = func(value reflect.Value, v interface{}) (err error) {
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
				return fallbackSetter(value, v, field.Set)
			}
			return err
		}
	default:
		field.Set = func(value reflect.Value, v interface{}) (err error) {
			return fallbackSetter(value, v, field.Set)
		}
	}
}
