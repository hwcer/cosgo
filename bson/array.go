package bson

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
	"strconv"
)

type Array struct {
	raw   bsoncore.Array //EmbeddedDocument || Array
	dict  []*Element     //子对象
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
	dict := make([]*Element, len(values))
	for i, v := range values {
		k := strconv.Itoa(i)
		if dict[i], err = NewElement(k, v); err != nil {
			return err
		}
	}
	this.dict = dict
	return nil
}

func (this *Array) Value(key string) bsoncore.Value {
	if IsTop(key) {
		return bsoncore.Value{Data: this.Build(), Type: bsontype.Array}
	}
	ele := this.Element(key)
	if ele != nil {
		return ele.Value()
	}
	return bsoncore.Value{}
}

func (this *Array) Build() []byte {
	if !this.dirty {
		return this.raw
	}
	var l int = 1
	var values []bsoncore.Value
	for pos, e := range this.dict {
		l += 2
		v := e.build()
		l += len(strconv.Itoa(pos))
		l += len(v.Data)
		values = append(values, v)
	}
	b := make([]byte, 0, l)
	this.raw = bsoncore.BuildArray(b, values...)
	this.dirty = false
	return this.raw
}

func (this *Array) String() string {
	raw := this.Build()
	return bsoncore.Array(raw).String()
}

// Element 获取子对象
func (this *Array) Element(key string) (r *Element) {
	k1, k2 := Split(key)
	if k1 != "" {
		if pos, err := strconv.Atoi(k1); err == nil && pos >= 0 && pos < len(this.dict) {
			r = this.dict[pos]
		}
	}
	if r != nil && k2 != "" {
		r = r.Element(k2)
	}
	return
}
func (this *Array) Marshal(key string, val interface{}) (n int, err error) {
	k1, k2 := Split(key)
	if k1 == "" {
		return this.marshal(val)
	}
	var pos int
	if pos, err = strconv.Atoi(k1); err != nil {
		return
	}
	if pos < 0 || pos >= len(this.dict) {
		return //TODO自动创建？
	}
	ele := this.dict[pos]
	n, err = ele.Marshal(k2, val)
	if err == nil && n != 0 {
		this.dirty = true
	}
	return
}

func (this *Array) Unmarshal(key string, i interface{}) (err error) {
	k1, k2 := Split(key)
	if k1 == "" {
		return this.unmarshal(i)
	}
	var pos int
	if pos, err = strconv.Atoi(k1); err != nil {
		return
	}
	if pos < 0 || pos >= len(this.dict) {
		return
	}
	ele := this.dict[pos]
	err = ele.Unmarshal(k2, i)
	return
}

// marshal 编译子对象  r字节变化量
func (this *Array) marshal(i interface{}) (r int, err error) {
	t, b, err := bson.MarshalValue(i)
	if err != nil {
		return
	}
	r = len(b) - len(this.raw)
	if t != bsontype.Array {
		r = -1
	}
	if r == 0 {
		this.raw = append(this.raw[0:0], b...)
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
