package bson

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type Element struct {
	key string
	raw bsoncore.Value
	arr *Array    //仅当type == bsontype.Array 时有效
	doc *Document //仅当type == bsontype.EmbeddedDocument 时有效
}

func (this *Element) ParseElement() (err error) {
	if this.Type() == bsontype.Array {
		this.arr, err = NewArray(this.raw.Data)
	} else if this.Type() == bsontype.EmbeddedDocument {
		this.doc, err = New(this.raw.Data)
	}
	return
}

func (this *Element) Key() string {
	return this.key
}

func (this *Element) Type() bsontype.Type {
	return this.raw.Type
}

// Value 仅当Element为EmbeddedDocument时才可以使用 key
func (this *Element) Value() bsoncore.Value {
	return this.build()
}
func (this *Element) GetBool() (r bool) {
	v := this.Value()
	r, _ = v.BooleanOK()
	return
}

func (this *Element) GetInt32() (r int32) {
	v := this.Value()
	r, _ = v.AsInt32OK()
	return
}

func (this *Element) GetInt64() (r int64) {
	v := this.Value()
	r, _ = v.AsInt64OK()
	return
}

func (this *Element) GetFloat() (r float64) {
	v := this.Value()
	r, _ = v.AsFloat64OK()
	return
}

func (this *Element) GetString() (r string) {
	v := this.Value()
	r, _ = v.StringValueOK()
	return
}

func (this *Element) String() string {
	raw := this.Build()
	return bsoncore.Element(raw).String()
}

func (this *Element) Build() []byte {
	k := this.Key()
	v := this.build()
	b := make([]byte, 0, len(v.Data)+len(k)+2)
	return bsoncore.AppendValueElement(b, k, v)
}

func (this *Element) Element(key string) *Element {
	if IsTop(key) {
		return this
	}
	if this.Type() == bsontype.Array {
		return this.arr.Element(key)
	} else if this.Type() == bsontype.EmbeddedDocument {
		return this.doc.Element(key)
	} else {
		return nil
	}
}

func (this *Element) Marshal(key string, i interface{}) (r int, err error) {
	if IsTop(key) {
		return this.marshal(i)
	} else if this.Type() == bsontype.Array {
		return this.arr.Marshal(key, i)
	} else if this.Type() == bsontype.EmbeddedDocument {
		return this.doc.Marshal(key, i)
	} else {
		return 0, fmt.Errorf("element type error:%v", key)
	}
}

func (this *Element) Unmarshal(key string, i interface{}) error {
	if IsTop(key) {
		return this.unmarshal(i)
	} else if this.Type() == bsontype.Array {
		return this.arr.Unmarshal(key, i)
	} else if this.Type() == bsontype.EmbeddedDocument {
		return this.doc.Unmarshal(key, i)
	} else {
		return nil
	}
}

func (this *Element) build() bsoncore.Value {
	if this.Type() == bsontype.Array && this.arr.dirty {
		this.raw.Data = this.arr.Build()
	} else if this.Type() == bsontype.EmbeddedDocument && this.doc.dirty {
		this.raw.Data = this.doc.Bytes()
	}
	return this.raw
}

// marshal  r字节变化量
func (this *Element) marshal(i interface{}) (r int, err error) {
	if this.Type() == bsontype.Array {
		return this.arr.marshal(i)
	} else if this.Type() == bsontype.EmbeddedDocument {
		return this.doc.marshal(i)
	}
	t, b, err := bson.MarshalValue(i)
	if err != nil {
		return
	}
	r = len(b) - len(this.raw.Data)
	if t != this.Type() {
		r = -1
	}
	if r == 0 {
		this.raw.Data = append(this.raw.Data[0:0], b...)
	} else {
		this.raw = bsoncore.Value{Type: t, Data: b}
	}
	return
}

func (this *Element) unmarshal(i interface{}) error {
	raw := bson.RawValue{Value: this.raw.Data, Type: this.raw.Type}
	return raw.Unmarshal(i)
}
