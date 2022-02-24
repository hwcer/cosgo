package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	*mongo.Database
}

func (this *Database) Collection(name string, opts ...*options.CollectionOptions) *Collection {
	coll := &Collection{}
	coll.Collection = this.Database.Collection(name, opts...)
	return coll
}
