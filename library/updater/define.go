package updater

import (
	"github.com/hwcer/cosgo/storage/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

/*
UGet 统一返回[]bson.M
*/
type ActType uint8

const (
	ActTypeAdd      ActType = 1 //添加
	ActTypeSub              = 2 //扣除
	ActTypeSet              = 3 //set
	ActTypeDel              = 4 //del
	ActTypeNew              = 5 //新对象
	ActTypeResolve          = 6 //自动分解
	ActTypeOverflow         = 7 //道具已满使用其他方式(邮件)转发
)

const (
	ItemNameId  = "id" //ITEM 对象中的ITEM Index 字段
	ItemNameVal = "val"
)

type updaterListenerType int32

const (
	UpdaterListenerTypeBeforeData updaterListenerType = iota
	UpdaterListenerTypeFinishData
	UpdaterListenerTypeBeforeSave
	UpdaterListenerTypeFinishSave
	UpdaterListenerTypeAdd
	UpdaterListenerTypeSub
)

type Cache struct {
	ID string      `json:"_id"`
	Id int32       `json:"id"`
	T  ActType     `json:"t"`
	K  string      `json:"k"`
	V  interface{} `json:"v"`
	B  int32       `json:"b"`
	R  interface{} `json:"r"`
}

//统一释放
type release interface {
	release()
}

//道具自动分解
type resolve interface {
	Resolve(id int32, num int32) (newId int32, newNum int32, ok bool)
}

type Worker interface {
	Add(id int32, num int64)
	Sub(id int32, num int64)
	Set(id interface{}, val interface{})
	Val(id int32) int64
	Keys(...interface{})
	Data() error
	Sort() int32
	Save() ([]*Cache, error)
	Verify() error
	Fields(...string)
}

type modelHash interface {
	USet(oid string, update mongo.Update) error     //使用主键更新
	UGet(oid string, keys []string) (bson.M, error) //使用主键初始化数据
	ObjectId(uid string, now time.Time) (oid string, err error)
}

type modelTable interface {
	New(uid string, iid int32, val int64, bag int32) (interface{}, error) //新对象
	UGet(uid string, query mongo.Query) ([]interface{}, error)            //使用主键初始化数据
	Parse(oid string) (iid int32, err error)
	ObjectId(uid string, iid int32) (oid string, err error)
	BulkWrite() *mongo.BulkWrite
}

type modelSetOnInert interface {
	SetOnInert(uid string, stime time.Time) bson.M
}

type itemHMap interface {
	Add(v int64) int64
	Sub(v int64) int64
	Set(map[string]interface{})
	GetOId() string
	GetIId() int32
	GetVal() int64
	Clone() interface{}
}

type binder interface {
	Keys(keys ...interface{})
	Fields(keys ...string)
	Dataset() *Dataset
	BulkWrite() *mongo.BulkWrite
}
