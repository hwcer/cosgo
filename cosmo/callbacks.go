package cosmo

import (
	"github.com/hwcer/cosgo/cosmo/clause"
	"reflect"
)

func initializeCallbacks() *callbacks {
	cb := &callbacks{processors: make(map[string]*processor)}
	cb.processors["query"] = &processor{handle: cmdQuery}
	cb.processors["create"] = &processor{handle: cmdCreate}
	cb.processors["update"] = &processor{handle: cmdUpdate}
	cb.processors["delete"] = &processor{handle: cmdDelete}
	return cb
}

// callbacks gorm callbacks manager
type callbacks struct {
	processors map[string]*processor
}

type processor struct {
	handle ExecuteHandle
}

//Call 自定义调用
func (cs *callbacks) Call(db *DB, handle ExecuteHandle) *DB {
	p := &processor{handle: handle}
	return p.Execute(db)
}

func (cs *callbacks) Create() *processor {
	return cs.processors["create"]
}

func (cs *callbacks) Query() *processor {
	return cs.processors["query"]
}

func (cs *callbacks) Update() *processor {
	return cs.processors["update"]
}

func (cs *callbacks) Delete() *processor {
	return cs.processors["delete"]
}

//Execute 执行操作
//  handle func(tx *DB,query bson.M) error
func (p *processor) Execute(db *DB) (tx *DB) {
	tx = db.Statement.Parse()
	if tx.Error != nil {
		return
	}
	stmt := tx.Statement
	//dest || model 类型为Struct并且主键不为空时，设置为查询条件
	var reflectModel reflect.Value
	if stmt.Model != nil {
		reflectModel = reflect.Indirect(stmt.ReflectModel)
	} else if stmt.ReflectValue.IsValid() && stmt.ReflectValue.Kind() == reflect.Struct {
		reflectModel = reflect.Indirect(stmt.ReflectValue)
	}
	if reflectModel.IsValid() {
		field := stmt.Schema.LookUpField(clause.MongoPrimaryName)
		if field != nil {
			v := reflectModel.FieldByIndex(field.StructField.Index)
			if v.IsValid() && !v.IsZero() {
				stmt.Clause.Primary(v.Interface())
			}
		}
	}

	if p.handle == nil || tx.Error != nil {
		return
	}
	defer func() {
		tx.Statement.Model = nil
		tx.Statement.Schema = nil
		tx.Statement.Paging = &Paging{}
		tx.Statement.Clause = clause.New()
		tx.Statement.multiple = false
	}()
	if err := p.handle(tx); err != nil {
		tx.Errorf(err)
		return
	}
	//fmt.Printf("Execute:%v,%+v\n", stmt.ReflectValue.Kind(), stmt.ReflectValue.Interface())
	return
}
