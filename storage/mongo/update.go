package mongo

import (
	"encoding/json"
	"fmt"
	"github.com/hwcer/cosgo/library/structs"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
)

type Update bson.M

func (this Update) v2bson(k string) bson.M {
	if v, ok := this[k]; !ok {
		return nil
	} else if m, ok2 := v.(bson.M); ok2 {
		return m
	}
	return nil
}

// STRUCT 通过STRUCT设置匹配参数
func (this Update) STRUCT(i interface{}) (id interface{}, err error) {
	if !structs.IsStruct(i) {
		return nil, fmt.Errorf("NOT STRUCT:%v", i)
	}
	s := structs.New(i)
	s.TagName = MongoTagName
	m := s.Map()
	for k, v := range m {
		if k != MongoPrimarykey {
			this[k] = v
		} else {
			id = v
		}
	}
	return
}

func (this Update) Set(k string, v interface{}) {
	this.Any("$set", k, v)
}

func (this Update) SetOnInert(k string, v interface{}) {
	this.Any("$setOnInsert", k, v)
}

func (this Update) Inc(k string, v interface{}) {
	this.Any("$inc", k, v)
}

//func (this Update) IncInt(k string, v int64) {
//	this.any("$inc", k, v)
//}
//func (this Update) IncFloat(k string, v float64) {
//	this.any("$inc", k, v)
//}

func (this Update) Min(k string, v interface{}) {
	this.Any("$min", k, v)
}

func (this Update) Max(k string, v interface{}) {
	this.Any("$max", k, v)
}

func (this Update) Unset(k string, v interface{}) {
	this.Any("$unset", k, v)
}

func (this Update) Pop(k string, v interface{}) {
	this.Any("$pop", k, v)
}

func (this Update) Pull(k string, v interface{}) {
	this.Any("$pull", k, v)
}

func (this Update) Push(k string, v interface{}) {
	this.Any("$push", k, v)
}

func (this Update) Any(t, k string, v interface{}) {
	if !strings.HasPrefix(t, "$") {
		t = "$" + t
	}
	if b := this.v2bson(t); b != nil {
		b[k] = v
	} else {
		this[t] = bson.M{k: v}
	}
}

func (this Update) String() string {
	b, _ := json.Marshal(this)
	return string(b)
}

func (this *Update) Clear() {
	*this = make(Update)
}
