package cosmo

import (
	"github.com/hwcer/cosgo/cosmo/clause"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
)

// Create insert the value into dbname
func cmdCreate(tx *DB) (err error) {
	coll := tx.client.Database(tx.dbname).Collection(tx.Statement.Table)
	switch tx.Statement.ReflectValue.Kind() {
	case reflect.Map, reflect.Struct:
		opts := options.InsertOne()
		if _, err = coll.InsertOne(tx.Statement.Context, tx.Statement.Dest, opts); err == nil {
			tx.RowsAffected = 1
		}
	case reflect.Array, reflect.Slice:
		opts := options.InsertMany()
		var documents []interface{}
		for i := 0; i < tx.Statement.ReflectValue.Len(); i++ {
			documents = append(documents, tx.Statement.ReflectValue.Index(i).Interface())
		}
		var result *mongo.InsertManyResult
		if result, err = coll.InsertMany(tx.Statement.Context, documents, opts); err == nil {
			tx.RowsAffected = int64(len(result.InsertedIDs))
		}
	}
	return
}

//Update 通用更新
// map ,bson.m 支持 $set $incr $setOnInsert, 其他未使用$字段一律视为$set操作
//支持struct 保存所有非零值
func cmdUpdate(tx *DB) (err error) {
	var update bson.M
	if update, err = tx.bson(tx.Statement.Dest); err != nil {
		return
	}
	//fmt.Printf("update:%+v\n", update)
	filter := tx.Statement.Clause.Build()
	if len(filter) == 0 {
		return ErrMissingWhereClause
	}
	//fmt.Printf("Update filter:%+v\n", filter)
	coll := tx.client.Database(tx.dbname).Collection(tx.Statement.Table)
	values := make(map[string]interface{})
	if tx.Statement.multiple || clause.Multiple(filter) {
		opts := options.Update()
		var result *mongo.UpdateResult
		if result, err = coll.UpdateMany(tx.Statement.Context, filter, update, opts); err == nil {
			tx.RowsAffected = result.MatchedCount
		}
	} else if tx.Statement.Model != nil && tx.Statement.ReflectModel.Kind() == reflect.Ptr {
		opts := options.FindOneAndUpdate()
		if _, ok := update[MongoSetOnInsert]; ok {
			opts.SetUpsert(true)
		}
		opts.SetReturnDocument(options.After)

		if projection, _, _ := tx.Statement.projection(); len(projection) > 0 {
			opts.SetProjection(projection)
		}
		if updateResult := coll.FindOneAndUpdate(tx.Statement.Context, filter, update, opts); updateResult.Err() == nil {
			tx.RowsAffected = 1
			err = updateResult.Decode(&values)
		} else {
			err = updateResult.Err()
		}
	} else {
		opts := options.Update()
		if _, ok := update[MongoSetOnInsert]; ok {
			opts.SetUpsert(true)
		}
		var result *mongo.UpdateResult
		if result, err = coll.UpdateOne(tx.Statement.Context, filter, update, opts); err == nil {
			tx.RowsAffected = result.MatchedCount
		}
	}

	if err != nil {
		tx.Error = err
		return
	}

	if len(values) > 0 {
		for k, v := range values {
			_ = tx.Statement.SetColumn(k, v)
		}
	}
	return
}

// Delete delete value match given conditions, if the value has primary key, then will including the primary key as condition
func cmdDelete(tx *DB) (err error) {
	filter := tx.Statement.Clause.Build()
	if len(filter) == 0 {
		return ErrMissingWhereClause
	}
	coll := tx.client.Database(tx.dbname).Collection(tx.Statement.Table)
	var result *mongo.DeleteResult
	if clause.Multiple(filter) {
		result, err = coll.DeleteMany(tx.Statement.Context, filter)
	} else {
		result, err = coll.DeleteOne(tx.Statement.Context, filter)
	}
	if err == nil {
		tx.RowsAffected = result.DeletedCount
	}
	return
}

//Find find records that match given conditions
//dest must be a pointer to a slice
func cmdQuery(tx *DB) (err error) {
	filter := tx.Statement.Clause.Build()

	//b, _ := json.Marshal(filter)
	//fmt.Printf("Query Filter:%+v\n", string(b))
	var multiple bool
	switch tx.Statement.ReflectValue.Kind() {
	case reflect.Array, reflect.Slice:
		multiple = true
	default:
		multiple = false
	}
	var (
		order      bson.D
		projection bson.M
	)
	if projection, order, err = tx.Statement.projection(); err != nil {
		return
	}

	coll := tx.client.Database(tx.dbname).Collection(tx.Statement.Table)
	if !multiple {
		opts := options.FindOne()
		if tx.Statement.Paging.offset > 0 {
			opts.SetSkip(int64(tx.Statement.Paging.offset))
		}
		if len(order) > 0 {
			opts.SetSort(order)
		}
		if len(projection) > 0 {
			opts.SetProjection(projection)
		}
		if result := coll.FindOne(tx.Statement.Context, filter, opts); result.Err() == nil {
			if err = result.Decode(tx.Statement.Dest); err == nil {
				tx.RowsAffected = 1
			}
		} else if result.Err() != mongo.ErrNoDocuments {
			err = result.Err()
		}
	} else {
		opts := options.Find()
		if tx.Statement.Paging.limit > 0 {
			opts.SetLimit(int64(tx.Statement.Paging.limit))
		}
		if tx.Statement.Paging.offset > 0 {
			opts.SetSkip(int64(tx.Statement.Paging.offset))
		}
		if len(order) > 0 {
			opts.SetSort(order)
		}
		if len(projection) > 0 {
			opts.SetProjection(projection)
		}
		var cursor *mongo.Cursor
		if cursor, err = coll.Find(tx.Statement.Context, filter, opts); err != nil {
			return
		}
		cursor.RemainingBatchLength()
		if err = cursor.All(tx.Statement.Context, tx.Statement.Dest); err == nil {
			tx.RowsAffected = int64(tx.Statement.ReflectValue.Len())
		}
	}

	return
}
