package cosmo

// Model specify the model you would like to run db operations
//    // update all users's name to `hello`
//    db.Model(&User{}).Update("name", "hello")
//    // if user's primary key is non-blank, will use it as condition, then will only update the user's name to `hello`
//    db.Model(&user).Update("name", "hello")
func (db *DB) Model(value interface{}) (tx *DB) {
	tx = db.getInstance(true)
	tx.Statement.Model = value
	return
}

// Table specify the table you would like to run db operations
//使用TABLE 时select,order,Omit 中必须使用数据库字段名称
func (db *DB) Table(name string) (tx *DB) {
	tx = db.getInstance(true)
	tx.Statement.Table = name
	return
}

// Multiple 强制批量操作
func (db *DB) Multiple() (tx *DB) {
	tx = db.getInstance()
	tx.Statement.multiple = true
	return
}

// Select specify fields that you want when querying, creating, updating
func (db *DB) Select(columns ...string) (tx *DB) {
	tx = db.getInstance()
	if len(tx.Statement.Omits) > 0 {
		tx.Error = ErrSelectOnOmitsExist
		return
	}
	tx.Statement.Selects = columns
	return
}

// Omit specify fields that you want to ignore when creating, updating and querying
func (db *DB) Omit(columns ...string) (tx *DB) {
	tx = db.getInstance()
	if len(tx.Statement.Omits) > 0 {
		tx.Error = ErrOmitOnSelectsExist
		return
	}
	tx.Statement.Omits = columns
	return
}

// Where 查询条件
//参考 query包
func (db *DB) Where(query interface{}, args ...interface{}) (tx *DB) {
	tx = db.getInstance()
	tx.Statement.Clause.Where(query, args...)
	return
}

//
//// Not add NOT conditions
func (db *DB) Not(query interface{}) (tx *DB) {
	tx = db.getInstance()
	tx.Statement.Clause.NOT(query)
	return
}

//
//// Or add OR conditions
func (db *DB) Or(query interface{}) (tx *DB) {
	tx = db.getInstance()
	tx.Statement.Clause.OR(query)
	return
}

// Page 分页设置 page-当前页，size-每页大小
func (db *DB) Page(page, size int) (tx *DB) {
	tx = db.getInstance()
	tx.Statement.Paging.Page(page, size)
	return
}

// Order specify order when retrieve records from dbname
func (db *DB) Order(key string, value int) (tx *DB) {
	tx = db.getInstance()
	tx.Statement.Paging.Order(key, value)
	return
}

// Limit specify the number of records to be retrieved
func (db *DB) Limit(limit int) (tx *DB) {
	tx = db.getInstance()
	tx.Statement.Paging.Limit(limit)
	return
}

// Offset specify the number of records to skip before starting to return the records
func (db *DB) Offset(offset int) (tx *DB) {
	tx = db.getInstance()
	tx.Statement.Paging.Offset(offset)
	return
}
