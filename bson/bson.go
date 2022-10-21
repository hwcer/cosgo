package bson

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
	"strings"
)

func New(raw []byte) (doc *Document, err error) {
	doc = &Document{raw: raw}
	err = doc.ParseElement()
	return
}

func NewArray(raw []byte) (arr *Array, err error) {
	arr = &Array{raw: raw}
	err = arr.ParseElement()
	return
}

func NewElement(key string, val bsoncore.Value) (ele *Element, err error) {
	ele = &Element{raw: val, key: key}
	if err = ele.raw.Validate(); err != nil {
		return
	}
	err = ele.ParseElement()
	return
}

func Marshal(v interface{}) (*Document, error) {
	raw, err := bson.Marshal(v)
	if err != nil {
		return nil, err
	}
	return New(raw)
}

func Unmarshal(doc *Document, v interface{}) error {
	return bson.Unmarshal(doc.raw, v)
}

func IsTop(k string) bool {
	if k == "" || k == "." {
		return true
	}
	return false
}

func Split(key string) (string, string) {
	if IsTop(key) {
		return "", ""
	}
	idx := strings.Index(key, ".")
	if idx < 0 {
		return key, ""
	} else {
		return key[0:idx], key[idx+1:]
	}
}
