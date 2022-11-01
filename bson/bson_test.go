package bson

import (
	"github.com/hwcer/logger"
	"testing"
)

type player struct {
	Id    string
	Name  string
	Lv    int32
	Info  Info
	Items []int32
}

type Info struct {
	Vip  int32
	Desc string
}

func TestBson(t *testing.T) {
	p := &player{
		Id:    "1",
		Name:  "hwc",
		Lv:    100,
		Items: []int32{1, 2},
	}
	doc, err := Marshal(p)
	if err != nil {
		logger.Debug(" err: %v", err)
		return
	}
	//var err error
	for _, k := range doc.Keys() {
		ele := doc.Element(k)
		logger.Debug("ele: %v", ele.String())
	}

	logger.Debug("Get top String:%v", doc.GetString("name"))

	_ = doc.Set("info", Info{Vip: 10})
	_ = doc.Set("items", []int32{1, 2, 3, 4})
	if err = doc.Set("info.vip", 100); err != nil {
		logger.Debug("error:%v", err)
	}
	logger.Debug("Get Struct int32:%v", doc.GetInt32("info.vip"))

	if err = doc.Set("name", "yyds"); err != nil {
		logger.Debug("error:%v", err)
	}

	if err = doc.Set("items.2", 200); err != nil {
		logger.Debug("error:%v", err)
	}

	logger.Debug("Get Array int32:%v", doc.GetInt32("items.2"))

	logger.Debug("new Document:%v", doc.String())

	b := []byte("abc")
	if v, e := Marshal(b); e != nil {
		logger.Debug("error:%v", e)
	} else {
		logger.Debug("success:%v", v)
	}

}
