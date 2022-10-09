package bson

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type Document struct {
	raw   bsoncore.Document
	dict  map[string]*Element //子对象
	dirty bool
}

func (this *Document) ParseElement() error {
	if err := this.raw.Validate(); err != nil {
		return err
	}
	arr, err := this.raw.Elements()
	if err != nil {
		return err
	}
	dict := make(map[string]*Element, len(arr))
	for _, ele := range arr {
		k := ele.Key()
		if dict[k], err = NewElement(k, ele.Value()); err != nil {
			return err
		}
	}
	this.dict = dict
	return nil
}

func (this *Document) Keys() (r []string) {
	for _, e := range this.dict {
		r = append(r, e.Key())
	}
	return
}

func (this *Document) Value(key string) bsoncore.Value {
	if IsTop(key) {
		return bsoncore.Value{Data: this.raw, Type: bsontype.EmbeddedDocument}
	}
	ele := this.Element(key)
	if ele != nil {
		return ele.Value()
	}
	return bsoncore.Value{}
}

func (this *Document) GetInt32(key string) (r int32) {
	v := this.Value(key)
	r, _ = v.Int32OK()
	return
}

func (this *Document) GetInt64(key string) (r int64) {
	v := this.Value(key)
	r, _ = v.Int64OK()
	return
}

func (this *Document) GetFloat(key string) (r float64) {
	v := this.Value(key)
	r, _ = v.AsFloat64OK()
	return
}
func (this *Document) GetBool(key string) (r bool) {
	v := this.Value(key)
	r, _ = v.BooleanOK()
	return
}

func (this *Document) GetString(key string) (r string) {
	v := this.Value(key)
	r, _ = v.StringValueOK()
	return
}

func (this *Document) Set(key string, val interface{}) error {
	_, err := this.Marshal(key, val)
	return err
}

func (this *Document) Build() []byte {
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

func (this *Document) String() string {
	raw := this.Build()
	return bsoncore.Document(raw).String()
}

// Element 获取子对象
func (this *Document) Element(key string) (r *Element) {
	k1, k2 := Split(key)
	if k1 != "" {
		r = this.dict[k1]
	}
	if r != nil && k2 != "" {
		r = r.Element(k2)
	}
	return
}
func (this *Document) Marshal(key string, val interface{}) (n int, err error) {
	k1, k2 := Split(key)
	if k1 == "" {
		return this.marshal(val)
	}
	ele := this.dict[k1]
	if ele == nil {
		err = bsoncore.ErrElementNotFound //TODO自动创建？
	} else {
		n, err = ele.Marshal(k2, val)
	}
	if err == nil && n != 0 {
		this.dirty = true
	}
	return
}

func (this *Document) Unmarshal(key string, i interface{}) (err error) {
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
func (this *Document) marshal(i interface{}) (r int, err error) {
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

func (this *Document) unmarshal(i interface{}) error {
	raw := this.Build()
	return bson.Unmarshal(raw, i)
}
