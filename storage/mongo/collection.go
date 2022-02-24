package mongo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//Collection Collection的封装
type Collection struct {
	*mongo.Collection
	Upsert bool
}

func KeysToProjection(keys interface{}) (r bson.M) {
	switch keys.(type) {
	case map[string]interface{}:
		r = keys.(map[string]interface{})
	case bson.M:
		r = keys.(bson.M)
	case []string:
		r = bson.M{}
		for _, k := range keys.([]string) {
			r[k] = 1
		}
	}
	return
}

//Get 主键查询(更新)一条记录，let(update)>0 更新记录
//keys   []string  ,bson.M ,map[string]interface{}   OID=-1,uid=1
//返回更新前的文档
func (this *Collection) Get(id interface{}, keys interface{}, result interface{}, update ...Update) (err error) {
	query := NewPrimaryQuery(id)
	projection := KeysToProjection(keys)
	if len(update) == 0 {
		opts := options.FindOne()
		if len(projection) > 0 {
			opts.SetProjection(projection)
		}
		err = this.Collection.FindOne(context.Background(), query, opts).Decode(result)
	} else {
		opts := options.FindOneAndUpdate()
		opts.SetUpsert(this.Upsert)
		if len(projection) > 0 {
			opts.SetProjection(projection)
		}
		err = this.Collection.FindOneAndUpdate(context.Background(), query, update[0], opts).Decode(result)
	}
	if err == mongo.ErrNoDocuments {
		err = nil
		result = nil
	}
	return err
}

//Set 主键更新(插入)一条记录
//id支持_id值,Clause,bson.Export,bson.D
func (this *Collection) Set(id interface{}, update Update) (err error) {
	opts := options.Update()
	opts.SetUpsert(this.Upsert)
	query := NewPrimaryQuery(id)
	_, err = this.Collection.UpdateOne(context.Background(), query, update, opts)
	return
}

//All 查询符合条件的所有记录
func (this *Collection) All(query Query, keys interface{}, results interface{}) (err error) {
	var cursor *mongo.Cursor
	opts := options.Find()
	projection := KeysToProjection(keys)
	if len(projection) > 0 {
		opts.SetProjection(projection)
	}
	cursor, err = this.Collection.Find(context.Background(), query, opts)
	if err != nil {
		return err
	}
	err = cursor.All(context.Background(), results)
	return err
}

/* Page 分页查询
参数 query：查询条件
参数 results:结果集,must be a pointer to a slice
返回值 count 当前查询条件下记录数
results 结果集中按照pageView.Sort设置的排序方式进行排序分页后的结果，但是results结果本身是无序的
*/
func (this *Collection) Page(query Query, pageView *Page, results interface{}) (count int64, err error) {
	var cursor *mongo.Cursor
	count, err = this.Collection.CountDocuments(context.Background(), query)
	if err != nil || count == 0 {
		return
	}
	cursor, err = this.Collection.Find(context.Background(), query, pageView.Options())
	if err != nil {
		return
	}
	err = cursor.All(context.Background(), results)
	if err != nil {
		return
	}
	return
}

//Save 使用STRUCT对象更新一条数据，使用主键匹配，所有字段被覆盖
func (this *Collection) Save(i interface{}, upsert ...bool) (result *mongo.UpdateResult, err error) {
	var id interface{}
	update := Update{}
	id, err = update.STRUCT(i)
	if err != nil {
		return
	}
	if id == nil {
		err = errors.New("mongo save not Primary key")
		return
	}
	opts := options.Update()
	if len(upsert) > 0 {
		opts.SetUpsert(upsert[0])
	} else {
		opts.SetUpsert(this.Upsert)
	}
	query := NewPrimaryQuery(id)
	return this.Collection.UpdateOne(context.Background(), query, update, opts)
}

//Update 更新多条记录
func (this *Collection) Update(query Query, update Update, multiple bool) (result *mongo.UpdateResult, err error) {
	opts := options.Update()
	if !multiple {
		if this.Upsert && update["$setOnInsert"] != nil {
			opts.SetUpsert(this.Upsert)
		}
		return this.Collection.UpdateOne(context.Background(), query, update, opts)
	} else {
		return this.Collection.UpdateMany(context.Background(), query, update, opts)
	}
}

func (this *Collection) Delete(query Query, multiple bool) (deleteCount int64, err error) {
	var result *mongo.DeleteResult
	if multiple {
		result, err = this.Collection.DeleteMany(context.Background(), query)
	} else {
		result, err = this.Collection.DeleteOne(context.Background(), query)
	}
	if err == nil {
		deleteCount = result.DeletedCount
	}
	return
}

func (this *Collection) Insert(documents ...interface{}) (InsertedID []interface{}, err error) {
	if len(documents) == 1 {
		opts := options.InsertOne()
		var result *mongo.InsertOneResult
		if result, err = this.Collection.InsertOne(context.Background(), documents[0], opts); err == nil {
			InsertedID = []interface{}{result.InsertedID}
		}
	} else {
		opts := options.InsertMany()
		var result *mongo.InsertManyResult
		if result, err = this.Collection.InsertMany(context.Background(), documents, opts); err == nil {
			InsertedID = result.InsertedIDs
		}
	}
	return
}

func (this *Collection) Count(query Query) (count int64, err error) {
	return this.Collection.CountDocuments(context.Background(), query)
}

//Indexes 使用原型管理索引
func (this *Collection) Indexes(indexes []Indexes) (err error) {
	if len(indexes) == 0 {
		return
	}
	indexView := this.Collection.Indexes()
	if len(indexes) == 1 {
		_, err = indexView.CreateOne(context.Background(), mongo.IndexModel(indexes[0]))
	} else {
		var arr []mongo.IndexModel
		for _, v := range indexes {
			arr = append(arr, mongo.IndexModel(v))
		}
		_, err = indexView.CreateMany(context.Background(), arr)
	}
	return
}

//Aggregate 聚合
func (this *Collection) Aggregate(pipeline interface{}, opts ...*options.AggregateOptions) (cursor *mongo.Cursor, err error) {
	return this.Collection.Aggregate(context.Background(), pipeline, opts...)
}

//批量写入
func (this *Collection) BulkWrite() *BulkWrite {
	return &BulkWrite{coll: this}
}
