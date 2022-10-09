package bson

import (
	"github.com/hwcer/logger"
	"testing"
)

type player struct {
	Id   string
	Name string
	Lv   int32
	Info Info
}

type Info struct {
	Vip  int32
	Desc string
}

func TestBson(t *testing.T) {
	p := &player{
		Id:   "1",
		Name: "hwc",
		Lv:   100,
	}
	doc, _ := Marshal(p)

	//var err error
	for _, k := range doc.Keys() {
		ele := doc.Element(k)
		logger.Debug(" ele, %v", ele.String())
	}

	logger.Debug("GetString --- name:%v", doc.GetString("name"))

	_ = doc.Set("info", Info{
		Vip:  10,
		Desc: "GOOD",
	})

	info := &Info{}
	_ = doc.Element("info").Unmarshal(info)
	logger.Debug("info:%+v", info)
	logger.Debug("Getint32 --- vip:%v", doc.GetInt32("info.vip"))
	//if err := doc.Set("info.vip", 100); err != nil {
	//	logger.Debug("error:%v", err)
	//}

	if err := doc.Set("name", "yyds"); err != nil {
		logger.Debug("error:%v", err)
	}

	logger.Debug("new Document:%v", doc.String())

}
