package cosmo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//分页
type Paging struct {
	order  []bson.E //排序
	limit  int      //每页大小
	offset int      //当前页
}

//Page 设置分页，page当前页,size每页大小
//相当于同时设置limit offset
func (this *Paging) Page(page, size int) {
	this.limit = size
	if page > 1 {
		this.offset = (page - 1) * size
	}
}

func (this *Paging) Limit(limit int) {
	this.limit = limit
}
func (this *Paging) Offset(offset int) {
	this.offset = offset
}

//Order 排序方式 1 和 -1 来指定排序的方式，其中 1 为升序排列，而 -1 是用于降序排列。
func (this *Paging) Order(key string, sort int) {
	if sort >= 0 {
		sort = 1
	} else {
		sort = -1
	}
	this.order = append(this.order, bson.E{
		Key: key, Value: sort,
	})
}

//Options 转换成FindOptions
func (this *Paging) Options() *options.FindOptions {
	opts := options.Find()
	opts.SetLimit(int64(this.limit))
	if this.offset > 1 {
		opts.SetSkip(int64(this.offset))
	}
	if len(this.order) > 0 {
		opts.SetSort(this.order)
	}
	return opts
}
