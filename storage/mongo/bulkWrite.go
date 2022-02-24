package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BulkWrite struct {
	coll   *Collection
	opts   []*options.BulkWriteOptions
	models []mongo.WriteModel
}

func (this *BulkWrite) Save() (result *mongo.BulkWriteResult, err error) {
	if len(this.models) == 0 {
		return nil, ErrorWriteModelEmpty
	}
	coll := this.coll.Collection
	if result, err = coll.BulkWrite(context.Background(), this.models, this.opts...); err == nil {
		this.models = nil
	}
	return
}

func (this *BulkWrite) Set(id interface{}, update Update) {
	model := mongo.NewUpdateOneModel()
	model.SetFilter(NewPrimaryQuery(id))
	model.SetUpdate(update)
	model.SetUpsert(this.coll.Upsert)
	this.models = append(this.models, model)
}

func (this *BulkWrite) Update(query Query, update Update) {
	model := mongo.NewUpdateManyModel()
	model.SetFilter(query)
	model.SetUpdate(update)
	model.SetUpsert(this.coll.Upsert)
	this.models = append(this.models, model)
}

func (this *BulkWrite) Insert(documents ...interface{}) {
	for _, doc := range documents {
		model := mongo.NewInsertOneModel()
		model.SetDocument(doc)
		this.models = append(this.models, model)
	}
}

func (this *BulkWrite) Delete(query Query, multiple bool) {
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
