package bson

import (
	"github.com/hwcer/logger"
	"go.mongodb.org/mongo-driver/bson"
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

func NewElement(raw []byte) (ele *Element, err error) {
	ele = &Element{raw: raw}
	if err = ele.raw.Validate(); err != nil {
		return
	}
	logger.Debug("NewElement %v:%v", ele.raw.Value().Type, ele.raw.String())
	err = ele.Parse()
	return
}

func Marshal(v interface{}) (*Document, error) {
	raw, err := bson.Marshal(v)
	if err != nil {
		return nil, err
	}
	return New(raw)
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
