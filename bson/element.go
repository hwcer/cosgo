package bson

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type Element struct {
	raw bsoncore.Element
	doc *Document //仅当type == bsontype.EmbeddedDocument 时有效
}

func (this *Element) Key() string {
	return this.raw.Key()
}

func (this *Element) Type() bsontype.Type {
	return this.raw.Value().Type
}

// Value 仅当Element为EmbeddedDocument时才可以使用 key
func (this *Element) Value() bsoncore.Value {
	_ = this.Build()
	return this.raw.Value()
}

func (this *Element) String() string {
	if this.Type() == bsontype.EmbeddedDocument {
		return this.doc.String()
	} else {
		return bsoncore.Element(this.Build()).String()
	}
}

func (this *Element) Build() []byte {
	if this.Type() == bsontype.EmbeddedDocument && this.doc.dirty {
		val := this.doc.Build()
		this.raw = bsoncore.AppendDocumentElement([]byte{}, this.Key(), val)
	}
	return this.raw
}
func (this *Element) Element(key string) *Element {
	if IsTop(key) {
		return this
	}
	if this.Type() != bsontype.EmbeddedDocument {
		return nil
	} else {
		return this.doc.Element(key)
	}
}

func (this *Element) Parse() (err error) {
	value := this.raw.Value()
	if value.Type == bsontype.Array {
		return this.parseEmbeddedArray(value)
	} else if value.Type == bsontype.EmbeddedDocument {
		return this.parseEmbeddedDocument(value)
	} else {
		return nil
	}
}

func (this *Element) parseEmbeddedArray(value bsoncore.Value) (err error) {
	_, ok := value.ArrayOK()
	if !ok {
		return bsoncore.ElementTypeError{Method: "bsoncore.Value.Document", Type: value.Type}
	}

	//this.doc, err = New(arr)
	return
}
func (this *Element) parseEmbeddedDocument(value bsoncore.Value) (err error) {
	doc, ok := value.DocumentOK()
	if !ok {
		return bsoncore.ElementTypeError{Method: "bsoncore.Value.Document", Type: value.Type}
	}
	this.doc, err = New(doc)
	return
}

func (this *Element) Marshal(key string, i interface{}) (r int, err error) {
	if IsTop(key) {
		return this.marshal(i)
	}
	if this.Type() != bsontype.EmbeddedDocument {
		return 0, fmt.Errorf("Element type not EmbeddedDocument :%v", key)
	}
	return this.doc.Marshal(key, i)
}

func (this *Element) Unmarshal(key string, i interface{}) error {
	if this.Type() == bsontype.EmbeddedDocument {
		return this.doc.Unmarshal(key, i)
	} else if IsTop(key) {
		return this.unmarshal(i)
	} else {
		return nil //TODO 报错？
	}
}

// marshal  r字节变化量
func (this *Element) marshal(i interface{}) (r int, err error) {
	if this.Type() == bsontype.EmbeddedDocument {
		return this.doc.marshal(i)
	}
	t, b, err := bson.MarshalValue(i)
	if err != nil {
		return
	}
	//if t != this.Type() {
	//	return 0, errors.New("type error")
	//}
	k := this.Key()
	r = len(b) - len(this.raw.Value().Data)
	v := bsoncore.Value{Type: t, Data: b}
	if r == 0 {
		this.raw = bsoncore.AppendValueElement(this.raw[0:0], k, v)
	} else {
		this.raw = bsoncore.AppendValueElement(make([]byte, 0, len(b)+len(k)+2), k, v)
	}
	return
}

func (this *Element) unmarshal(i interface{}) error {
	val := this.raw.Value()
	raw := bson.RawValue{Value: val.Data, Type: val.Type}
	return raw.Unmarshal(i)
}
