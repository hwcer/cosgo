package cosmo

//
//func (tx *DB) assignInterfacesToValue(values ...interface{}) {
//	for _, value := range values {
//		switch v := value.(type) {
//		case []clause.Expression:
//			for _, expr := range v {
//				if eq, ok := expr.(clause.Eq); ok {
//					switch column := eq.Column.(type) {
//					case string:
//						if field := tx.Statement.Schema.LookUpField(column); field != nil {
//							tx.Errorf(field.Set(tx.Statement.ReflectValue, eq.Value))
//						}
//					case clause.Column:
//						if field := tx.Statement.Schema.LookUpField(column.Name); field != nil {
//							tx.Errorf(field.Set(tx.Statement.ReflectValue, eq.Value))
//						}
//					}
//				} else if andCond, ok := expr.(clause.AndConditions); ok {
//					tx.assignInterfacesToValue(andCond.Exprs)
//				}
//			}
//		case clause.Expression, map[string]string, map[interface{}]interface{}, map[string]interface{}:
//			if exprs := tx.Statement.BuildCondition(value); len(exprs) > 0 {
//				tx.assignInterfacesToValue(exprs)
//			}
//		default:
//			if s, err := schema.ParseModel(value, tx.cacheStore, tx.NamingStrategy); err == nil {
//				reflectValue := reflect.Indirect(reflect.ValueOf(value))
//				switch reflectValue.Kind() {
//				case reflect.Struct:
//					for _, f := range s.Fields {
//						if f.Readable {
//							if v, isZero := f.ValueOf(reflectValue); !isZero {
//								if field := tx.Statement.Schema.LookUpField(f.Name); field != nil {
//									tx.Errorf(field.Set(tx.Statement.ReflectValue, v))
//								}
//							}
//						}
//					}
//				}
//			} else if len(values) > 0 {
//				if exprs := tx.Statement.BuildCondition(values[0], values[1:]...); len(exprs) > 0 {
//					tx.assignInterfacesToValue(exprs)
//				}
//				return
//			}
//		}
//	}
//}

//FirstOrInit 获取第一条匹配的记录，或者根据给定的条件初始化一个实例 （仅支持 sturct 和 map 条件）
// 仅仅初始化 dest 不会操作数据库
//
//func (db *DB) FirstOrInit(dest interface{}, conds ...interface{}) (tx *DB) {
//	queryTx := db.Limit(1)
//
//	if tx = queryTx.Find(dest, conds...); queryTx.RowsAffected == 0 {
//		if c, ok := tx.Statement.Clauses["WHERE"]; ok {
//			if where, ok := c.Expression.(clause.Where); ok {
//				tx.assignInterfacesToValue(where.Exprs)
//			}
//		}
//
//		// initialize with attrs, conds
//		if len(tx.Statement.attrs) > 0 {
//			tx.assignInterfacesToValue(tx.Statement.attrs...)
//		}
//	}
//
//	// initialize with attrs, conds
//	if len(tx.Statement.assigns) > 0 {
//		tx.assignInterfacesToValue(tx.Statement.assigns...)
//	}
//	return
//}
//
////FirstOrCreate 获取第一条匹配的记录，或者根据给定的条件创建一条新纪录（仅支持 sturct 和 map 条件）
////初始化dest 并且将结果写入到数据
//func (db *DB) FirstOrCreate(dest interface{}, conds ...interface{}) (tx *DB) {
//	queryTx := db.Limit(1).Order(clause.OrderByColumn{
//		Column: clause.Column{Table: clause.CurrentTable, Name: clause.PrimaryKey},
//	})
//	if tx = queryTx.Find(dest, conds...); tx.Error == nil {
//		if tx.RowsAffected == 0 {
//			if c, ok := tx.Statement.Clauses["WHERE"]; ok {
//				if where, ok := c.Expression.(clause.Where); ok {
//					tx.assignInterfacesToValue(where.Exprs)
//				}
//			}
//
//			// initialize with attrs, conds
//			if len(tx.Statement.attrs) > 0 {
//				tx.assignInterfacesToValue(tx.Statement.attrs...)
//			}
//
//			// initialize with attrs, conds
//			if len(tx.Statement.assigns) > 0 {
//				tx.assignInterfacesToValue(tx.Statement.assigns...)
//			}
//
//			return tx.Create(dest)
//		} else if len(db.Statement.assigns) > 0 {
//			exprs := tx.Statement.BuildCondition(db.Statement.assigns[0], db.Statement.assigns[1:]...)
//			assigns := map[string]interface{}{}
//			for _, expr := range exprs {
//				if eq, ok := expr.(clause.Eq); ok {
//					switch column := eq.Column.(type) {
//					case string:
//						assigns[column] = eq.Value
//					case clause.Column:
//						assigns[column.Name] = eq.Value
//					default:
//					}
//				}
//			}
//
//			return tx.Model(dest).Update(assigns)
//		}
//	}
//	return tx
//}
