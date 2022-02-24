package cosmo

import (
	"context"
	"errors"
	"fmt"
	"github.com/hwcer/cosgo/cosmo/schema"
	"github.com/hwcer/cosgo/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
)

// DB GORM DB definition
type DB struct {
	*Config
	clone        bool //是否克隆体
	Error        error
	Statement    *Statement
	RowsAffected int64 //操作影响的条数
}

// New
//address uri || *mongo.Client
func New(dbname string, configs ...*Config) (db *DB) {
	var config *Config
	if len(configs) > 0 {
		config = configs[0]
	} else {
		config = &Config{}
	}
	config.dbname = dbname

	if config.Schema == nil {
		config.Schema = schema.New(&sync.Map{}, &schema.NamingStrategy{})
	}

	if config.Logger == nil {
		config.Logger = logger.Default
	}

	if config.Plugins == nil {
		config.Plugins = map[string]Plugin{}
	}

	db = &DB{Config: config}
	db.callbacks = initializeCallbacks()
	db.Statement = NewStatement(db)

	return
}

func (db *DB) Start(address interface{}) (err error) {
	switch address.(type) {
	case string:
		db.Config.client, err = NewClient(address.(string))
	case *mongo.Client:
		db.Config.client = address.(*mongo.Client)
	default:
		err = errors.New("address error")
	}
	if err != nil {
		return
	}

	return
}

// Session create new db session
func (db *DB) Session(session *Session) *DB {
	var (
		config = *db.Config
		tx     = &DB{
			Config: &config,
			Error:  db.Error,
			clone:  false,
		}
	)

	tx.Statement = NewStatement(tx)

	if session.DBName != "" {
		tx.Config.dbname = session.DBName
	}

	if session.Context != nil {
		tx.Statement.Context = session.Context
	}

	if session.Logger != nil {
		tx.Config.Logger = config.Logger
	}

	return tx
}

//新数据库
func (db *DB) Database(dbname string) *DB {
	return db.Session(&Session{DBName: dbname})
}
func (db *DB) Collection(model interface{}) (tx *DB, coll *mongo.Collection) {
	switch model.(type) {
	case string:
		tx = db.Table(model.(string))
	default:
		tx = db.Model(model)
	}
	tx = tx.callbacks.Call(tx, func(tx *DB) error {
		coll = tx.client.Database(tx.dbname).Collection(tx.Statement.Table)
		return nil
	})
	return
}

//BulkWrite 批量写入
func (db *DB) BulkWrite(model interface{}) *BulkWrite {
	tx := db.Model(model)
	return &BulkWrite{tx: tx}
}

// WithContext change current instance db's context to ctx
func (db *DB) WithContext(ctx context.Context) *DB {
	return db.Session(&Session{Context: ctx})
}

// Set store value with key into current db instance's context
func (db *DB) Set(key string, value interface{}) *DB {
	tx := db.getInstance()
	setting := map[string]interface{}{}
	for k, v := range tx.Statement.settings {
		setting[k] = v
	}
	tx.Statement.settings = setting
	return tx
}

// Get get value with key from current db instance's context
func (db *DB) Get(key string) (val interface{}, ok bool) {
	val, ok = db.Statement.settings[key]
	return
}

// Errorf add error to db
func (db *DB) Errorf(msg interface{}, args ...interface{}) error {
	var err error
	switch msg.(type) {
	case string:
		err = fmt.Errorf(msg.(string), args...)
	case error:
		err = msg.(error)
	default:
		err = fmt.Errorf("%v", msg)
	}
	if db.Error == nil {
		db.Error = err
	} else if err != nil {
		db.Error = fmt.Errorf("%v; %w", db.Error, err)
	}
	return db.Error
}

func (db *DB) getInstance(clone ...bool) *DB {
	if len(clone) == 0 && db.clone {
		return db
	}
	tx := &DB{Config: db.Config, clone: true}
	tx.Statement = NewStatement(tx)
	return tx
}

func (db *DB) Use(plugin Plugin) error {
	name := plugin.Name()
	if _, ok := db.Plugins[name]; ok {
		return ErrRegistered
	}
	if err := plugin.Initialize(db); err != nil {
		return err
	}
	db.Plugins[name] = plugin
	return nil
}
