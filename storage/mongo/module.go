package mongo

import (
	"go.mongodb.org/mongo-driver/mongo/options"
)

type moduleTable struct {
	dbName     string
	collName   string
	indexes    []Indexes
	collection *Collection
}
type OnModuleStarted func(string, *Collection)

type Module struct {
	uri    string
	client *Client
	tables map[string]*moduleTable
}

func NewModule() *Module {
	return &Module{
		tables: make(map[string]*moduleTable),
	}
}

//Register 注册coll
func (this *Module) Register(dbName, collName string, indexes ...Indexes) {
	c := &moduleTable{
		dbName:   dbName,
		collName: collName,
		indexes:  indexes,
	}
	k := c.dbName + c.collName
	this.tables[k] = c
}

func (this *Module) Start(uri string, f OnModuleStarted) (err error) {
	if this.client, err = New(uri); err != nil {
		return
	}
	this.uri = uri
	for _, table := range this.tables {
		db := this.Database(table.dbName)
		table.collection = db.Collection(table.collName)
		if len(table.indexes) > 0 {
			table.collection.Indexes(table.indexes)
		}
		if f != nil {
			f(table.collName, table.collection)
		}
	}
	return
}

//Close 关闭时调用,breakOnError 遇到错误时停止关闭
func (this *Module) Close() (err error) {
	return this.client.Close()
}

func (this *Module) Client() *Client {
	return this.client
}

func (this *Module) Database(name string, opts ...*options.DatabaseOptions) *Database {
	return this.client.Database(name, opts...)
}

func (this *Module) Collection(name string) (*Collection, bool) {
	if table, ok := this.tables[name]; ok {
		return table.collection, true
	} else {
		return nil, false
	}
}
