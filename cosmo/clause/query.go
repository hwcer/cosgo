package clause

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
)

const (
	MongoPrimaryName     = "_id"
	QueryOperationPrefix = "$"
)

const (
	QueryOperationOR  = "or"
	QueryOperationAND = "and"
	QueryOperationNOT = "not"
	QueryOperationNOR = "nor"
)

const (
	UpdateTypeSet         = "$set"
	UpdateTypeInc         = "$inc"
	UpdateTypesetOnInsert = "$setOnInsert"
)

var complexCondition = []string{QueryOperationOR, QueryOperationAND, QueryOperationNOT, QueryOperationNOR}

func New() *Query {
	q := &Query{}
	q.complex = make(map[string][]interface{})
	for _, k := range complexCondition {
		q.complex[k] = []interface{}{}
	}
	return q
}

type Node struct {
	t string
	k string
	v []interface{}
}

type Query struct {
	where   []*Node
	primary []interface{} //主键
	complex map[string][]interface{}
}

//Primary 使用主键匹配 一个值或者数组
func (q *Query) Primary(id interface{}) {
	q.primary = append(q.primary, parseArray(id)...)
}

func (q *Query) any(t, k string, v ...interface{}) {
	if !strings.HasPrefix(t, QueryOperationPrefix) {
		t = QueryOperationPrefix + t
	}
	q.where = append(q.where, &Node{t: t, k: k, v: v})
}

//Match 特殊匹配 or,not,and,nor
func (this Query) match(t string, v interface{}) {
	t = strings.TrimPrefix(t, QueryOperationPrefix)
	r := parseArray(v)
	this.complex[t] = append(this.complex[t], r...)
}

//Eq 等于（=）
func (q *Query) Eq(k string, v interface{}) {
	q.any("", k, v)
}

//Gt 大于（>）
func (q *Query) Gt(k string, v interface{}) {
	q.any("$gt", k, v)
}

//Gte 大于等于（>=）
func (q *Query) Gte(k string, v interface{}) {
	q.any("$gte", k, v)
}

//Lt 小于（<）
func (q *Query) Lt(k string, v interface{}) {
	q.any("$lt", k, v)
}

//Lte 小于等于（<=）
func (q *Query) Lte(k string, v interface{}) {
	q.any("$lte", k, v)
}

//Ne 不等于（!=）
func (q *Query) Ne(k string, v interface{}) {
	q.any("$ne", k, v)
}

//In The $in operator selects the documents where the value of a field equals any value in the specified array
func (q *Query) In(k string, v interface{}) {
	q.any("$in", k, parseArray(v)...)
}

//Nin selects the documents where: the field value is not in the specified array or the field does not exist.
func (q *Query) Nin(k string, v interface{}) {
	q.any("$nin", k, parseArray(v)...)
}

//OR The $or operator performs a logical OR operation on an array of two or more <expressions> and selects the documents that satisfy at least one of the <expressions>.
func (this Query) OR(v interface{}) {
	this.match("or", v)
}

//NOT $not performs a logical NOT operation on the specified <operator-expression> and selects the documents that do not match the <operator-expression>.
//This includes documents that do not contain the field.
func (this Query) NOT(v interface{}) {
	this.match("not", v)
}

//AND $and performs a logical AND operation on an array of one or more expressions (e.g. <expression1>, <expression2>, etc.) and selects the documents that satisfy all the expressions in the array.
//The $and operator uses short-circuit evaluation. If the first expression (e.g. <expression1>) evaluates to false, MongoDB will not evaluate the remaining expressions.
func (this Query) AND(v interface{}) {
	this.match("and", v)
}

//NOR $nor performs a logical NOR operation on an array of one or more query expression and selects the documents that fail all the query expressions in the array.
func (this Query) NOR(v interface{}) {
	this.match("nor", v)
}

func (q *Query) Marshal() ([]byte, error) {
	return bson.Marshal(q.Build())
}

func (q *Query) String() string {
	b, _ := json.Marshal(q.Build())
	return string(b)
}
