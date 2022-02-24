package mongo

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
)

type Query bson.M

//主键查询
func NewPrimaryQuery(ids ...interface{}) Query {
	query := make(Query)
	query.Primary(ids...)
	return query
}

func (this Query) v2bson(k string) bson.M {
	if v, ok := this[k]; !ok {
		return nil
	} else if m, ok2 := v.(bson.M); ok2 {
		return m
	}
	return nil
}

//通过bson设置匹配参数
func (this Query) BSOND(d bson.D) {
	for _, v := range d {
		this[v.Key] = v.Value
	}
}

func (this Query) Primary(ids ...interface{}) {
	if len(ids) == 1 {
		this[MongoPrimarykey] = ids[0]
	} else {
		this.In(MongoPrimarykey, ids)
	}
}

func (this Query) PrimaryString(ids ...string) {
	if len(ids) == 1 {
		this[MongoPrimarykey] = ids[0]
	} else {
		this.In(MongoPrimarykey, ArrString(ids).Interface()...)
	}
}

//Match 特殊匹配 or,not,and,nor
func (this Query) Match(t string, query ...Query) {
	if !strings.HasPrefix(t, "$") {
		t = "$" + t
	}
	var old []Query
	if v, ok := this[t]; ok {
		if m, ok2 := v.([]Query); ok2 {
			old = m
		}
	}
	old = append(old, query...)
	this[t] = old
}

func (this Query) Any(t, k string, v interface{}) {
	if !strings.HasPrefix(t, "$") {
		t = "$" + t
	}
	if m := this.v2bson(k); m != nil {
		m[t] = v
	} else {
		m = make(bson.M)
		//m["$eq"] = this[k] //将原 {k:v} 模式转换成 {k:{"$gt":v}
		m[t] = v
		this[k] = m
	}
}

//Eq 等于（=）
func (this Query) Eq(k string, v interface{}) {
	this[k] = v
}

//Gt 大于（>）
func (this Query) Gt(k string, v interface{}) {
	this.Any("$gt", k, v)
}

//Gte 大于等于（>=）
func (this Query) Gte(k string, v interface{}) {
	this.Any("$gte", k, v)
}

//Lt 小于（<）
func (this Query) Lt(k string, v interface{}) {
	this.Any("$lt", k, v)
}

//Lte 小于等于（<=）
func (this Query) Lte(k string, v interface{}) {
	this.Any("$lte", k, v)
}

//Ne 不等于（!=）
func (this Query) Ne(k string, v interface{}) {
	this.Any("$ne", k, v)
}

//In The $in operator selects the documents where the value of a field equals any value in the specified array
func (this Query) In(k string, v ...interface{}) {
	this.Any("$in", k, v)
}

//Nin selects the documents where: the field value is not in the specified array or the field does not exist.
func (this Query) Nin(k string, v ...interface{}) {
	this.Any("$nin", k, v)
}

//OR The $or operator performs a logical OR operation on an array of two or more <expressions> and selects the documents that satisfy at least one of the <expressions>.
func (this Query) OR(v ...Query) {
	this.Match("$or", v...)
}

//NOT $not performs a logical NOT operation on the specified <operator-expression> and selects the documents that do not match the <operator-expression>.
//This includes documents that do not contain the field.
func (this Query) NOT(v ...Query) {
	this.Match("$not", v...)
}

//AND $and performs a logical AND operation on an array of one or more expressions (e.g. <expression1>, <expression2>, etc.) and selects the documents that satisfy all the expressions in the array.
//The $and operator uses short-circuit evaluation. If the first expression (e.g. <expression1>) evaluates to false, MongoDB will not evaluate the remaining expressions.
func (this Query) AND(v ...Query) {
	this.Match("$and", v...)
}

//NOR $nor performs a logical NOR operation on an array of one or more query expression and selects the documents that fail all the query expressions in the array.
func (this Query) NOR(v ...Query) {
	this.Match("$nor", v...)
}

func (this Query) Marshal() ([]byte, error) {
	return bson.Marshal(this)
}

func (this Query) String() string {
	b, _ := json.Marshal(this)
	return string(b)
}
