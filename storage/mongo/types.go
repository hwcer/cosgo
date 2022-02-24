package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
)

func NewPage(page, size int) *Page {
	return &Page{page: page, size: size}
}

//NewIndexes 创建索引,key: 字段列表，使用urlQuery格式；key1=sort1&key2=sort2...
//sort=1正序；-1 倒序
func NewIndexes(key string, unique bool) Indexes {
	indexes := Indexes{}
	var keys []bson.E
	for _, v := range strings.Split(key, "&") {
		arr := strings.Split(v, "=")
		var k = arr[0]
		sort := int(1)
		if len(arr) > 1 && arr[1] == "-1" {
			sort = -1
		}
		keys = append(keys, bson.E{Key: k, Value: sort})
	}

	indexes.Keys = keys
	indexes.Options = options.Index()
	indexes.Options.SetUnique(unique)
	return indexes
}

type Indexes mongo.IndexModel

//分页
type Page struct {
	page int    //当前页
	size int    //每页大小
	sort bson.D //排序
}

func (this *Page) Set(page, size int) {
	this.page = page
	this.size = size
}
func (this *Page) Sort(key string, sort int) {
	this.sort = append(this.sort, bson.E{
		Key: key, Value: sort,
	})
}

//Options 转换成FindOptions
func (this *Page) Options() *options.FindOptions {
	opts := options.Find()
	opts.SetLimit(int64(this.size))
	if this.page > 1 {
		skip := (this.page - 1) * this.size
		opts.SetSkip(int64(skip))
	}
	if len(this.sort) > 0 {
		opts.SetSort(this.sort)
	}
	return opts
}
