package bson

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
	"strconv"
	"strings"
)

type Array struct {
	raw   bsoncore.Array      //EmbeddedDocument || Array
	dict  map[string]*Element //子对象
	dirty bool
}

func (this *Array) ParseElement() error {
	if err := this.raw.Validate(); err != nil {
		return err
	}
	values, err := this.raw.Values()
	if err != nil {
		return err
	}
	dict := make(map[string]*Element, len(values))
	for i, ele := range values {
		k := strconv.Itoa(i)
		if dict[k], err = NewElement(ele); err != nil {
			return err
		}
	}
	this.dict = dict
	return nil
}

func (this *Array) Keys() (r []string) {
	for _, e := range this.dict {
		r = append(r, e.Key())
	}
	return
}

func (this *Array) Value(key string) bsoncore.Value {
	if IsTop(key) {
		return bsoncore.Value{Data: this.raw, Type: bsontype.EmbeddedDocument}
	}
	ele := this.Element(key)
	if ele != nil {
		return ele.Value()
	}
	return bsoncore.Value{}
}

func (this *Array) GetInt32(key string) (r int32) {
	v := this.Value(key)
	r, _ = v.Int32OK()
	return
}

func (this *Array) GetInt64(key string) (r int64) {
	v := this.Value(key)
	r, _ = v.Int64OK()
	return
}

func (this *Array) GetFloat(key string) (r float64) {
	v := this.Value(key)
	r, _ = v.AsFloat64OK()
	return
}
func (this *Array) GetBool(key string) (r bool) {
	v := this.Value(key)
	r, _ = v.BooleanOK()
	return
}

func (this *Array) GetString(key string) (r string) {
	v := this.Value(key)
	r, _ = v.StringValueOK()
	return
}

func (this *Array) Set(key string, val interface{}) error {
	_, err := this.Marshal(key, val)
	return err
}

func (this *Array) Build() []byte {
	if !this.dirty {
		return this.raw
	}
	var l int
	var ele [][]byte
	for _, e := range this.dict {
		b := e.Build()
		l += len(b)
		ele = append(ele, b)
	}
	raw := make([]byte, 0, l+4)
	this.raw = bsoncore.BuildDocument(raw, ele...)
	this.dirty = false
	return this.raw
}

func (this *Array) String() string {
	raw := this.Build()
	return bsoncore.Document(raw).String()
}

// Element 获取子对象
func (this *Array) Element(key string) *Element {
	idx := strings.Index(key, ".")
	if idx < 0 {
		return this.dict[key]
	}
	ele := this.dict[key[0:idx]]
	if ele != nil && ele.Type() == bsontype.EmbeddedDocument {
		return ele.Element(key[idx+1:])
	} else {
		return ele
	}
}
func (this *Array) Marshal(key string, val interface{}) (n int, err error) {
	k1, k2 := Split(key)
	if k1 == "" {
		n, err = this.marshal(val)
	} else if k2 == "" {
		ele := this.dict[k1]
		if ele == nil {
			err = bsoncore.ErrElementNotFound //TODO自动创建？
		} else {
			n, err = ele.Marshal("", val)
		}
	} else {
		ele := this.dict[k1]
		if ele == nil {
			err = bsoncore.ErrElementNotFound //TODO自动创建？
		} else {
			n, err = ele.Marshal(k2, val)
		}
	}
	if err == nil && n != 0 {
		this.dirty = true
	}
	return
}

func (this *Array) Unmarshal(key string, i interface{}) (err error) {
	k1, k2 := Split(key)
	if k1 == "" {
		err = this.unmarshal(i)
	} else if k2 == "" {
		ele := this.dict[k1]
		if ele == nil {
			err = bsoncore.ErrElementNotFound //TODO自动创建？
		} else {
			err = ele.Unmarshal("", i)
		}
	} else {
		ele := this.dict[k1]
		if ele == nil {
			err = bsoncore.ErrElementNotFound //TODO自动创建？
		} else {
			err = ele.Unmarshal(k2, i)
		}
	}
	return
}

// marshal 编译子对象  r字节变化量
func (this *Array) marshal(i interface{}) (r int, err error) {
	t, b, err := bson.MarshalValue(i)
	if err != nil {
		return
	}
	if t != bsontype.EmbeddedDocument {
		return 0, errors.New("type error")
	}
	r = len(b) - len(this.raw)
	if r == 0 {
		this.raw = bsoncore.AppendDocument(this.raw[0:0], b)
	} else {
		this.raw = b
		this.dirty = true
	}
	err = this.ParseElement()
	return
}

func (this *Array) unmarshal(i interface{}) error {
	raw := this.Build()
	return bson.Unmarshal(raw, i)
}
