package cosmo

import (
	"context"
	"errors"
	"github.com/hwcer/cosgo/cosmo/clause"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BulkWrite struct {
	tx     *DB
	opts   []*options.BulkWriteOptions
	models []mongo.WriteModel
}

func (this *BulkWrite) Save() (result *mongo.BulkWriteResult, err error) {
	if len(this.models) == 0 {
		return nil, errors.New("ErrorWriteModelEmpty")
	}
	tx := this.tx.callbacks.Call(this.tx, func(db *DB) error {
		coll := db.client.Database(db.dbname).Collection(db.Statement.Table)
		if result, err = coll.BulkWrite(context.Background(), this.models, this.opts...); err == nil {
			this.models = nil
		}
		return err
	})
	err = tx.Error
	return
}

func (this *BulkWrite) Update(data bson.M, where ...interface{}) {
	query := clause.New()
	query.Where(where[0], where[1:]...)
	update, _ := this.tx.parseMap(data)

	model := mongo.NewUpdateOneModel()
	model.SetFilter(query.Build())
	model.SetUpdate(update)
	if _, ok := update[clause.UpdateTypesetOnInsert]; ok {
		model.SetUpsert(true)
	}
	this.models = append(this.models, model)
}

func (this *BulkWrite) Insert(documents ...interface{}) {
	for _, doc := range documents {
		model := mongo.NewInsertOneModel()
		model.SetDocument(doc)
		this.models = append(this.models, model)
	}
}

func (this *BulkWrite) Delete(where ...interface{}) {
	query := clause.New()
	query.Where(where[0], where[1:]...)
	filter := query.Build()
	multiple := clause.Multiple(filter)

	if multiple {
		model := mongo.NewDeleteManyModel()
		model.SetFilter(query)
		this.models = append(this.models, model)
	} else {
		model := mongo.NewDeleteOneModel()
		model.SetFilter(query)
		this.models = append(this.models, model)
	}
}
