package cosmo

import (
	"context"
	"github.com/hwcer/cosgo/cosmo/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//AutoMigrator returns migrator
//Sparse
func (db *DB) AutoMigrator(dst ...interface{}) error {
	for _, mod := range dst {
		sch, err := db.Schema.Parse(mod)
		if err != nil {
			return err
		}
		indexes := sch.ParseIndexes()
		for _, index := range indexes {
			if e := db.indexes(mod, index); e != nil {
				return e
			}
		}
	}
	return nil
}

func (db *DB) indexes(model interface{}, index schema.Index) (err error) {
	indexes := mongo.IndexModel{}
	var keys []bson.E
	for _, field := range index.Fields {
		k := field.DBName
		v := 1
		if field.Sort == "desc" {
			v = -1
		}
		keys = append(keys, bson.E{Key: k, Value: v})
	}
	//fmt.Printf("index:%+v\n\n\n", index)
	indexes.Keys = keys
	indexes.Options = options.Index()
	indexes.Options.SetName(index.Name)
	if index.Unique {
		indexes.Options.SetUnique(true)
	}
	if index.Sparse {
		indexes.Options.SetSparse(true)
	}

	tx, coll := db.Collection(model)
	if tx.Error != nil {
		return tx.Error
	}
	indexView := coll.Indexes()
	_, err = indexView.CreateOne(context.Background(), indexes)

	return
}
