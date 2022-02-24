package clause

import (
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
)

var operationArray = []string{"$in", "$nin"}

//PrimaryClear 主键去重
func (q *Query) primaryClear() (r []interface{}) {
	dict := make(map[interface{}]bool)
	for _, k := range q.primary {
		if !dict[k] {
			dict[k] = true
			r = append(r, k)
		}
	}
	return
}

//Build 生成mongo查询条件
func (q *Query) Build() bson.M {
	query := make(bson.M)
	//defer fmt.Printf("Query Build:%+v\n", query)
	primary := q.primaryClear()
	if len(primary) == 1 {
		query[MongoPrimaryName] = primary[0]
		return query
	} else if len(primary) > 1 {
		query[MongoPrimaryName] = bson.M{"$in": primary}
	}
	for _, node := range q.where {
		build(query, node)
	}
	for _, k := range complexCondition {
		if len(q.complex[k]) > 0 {
			query[QueryOperationPrefix+k] = q.complex[k]
		}
	}

	return query
}

func isArray(node *Node) bool {
	for _, k := range operationArray {
		if k == node.t {
			return true
		}
	}
	return false
}
func parseArray(v interface{}) []interface{} {
	vf := reflect.Indirect(reflect.ValueOf(v))
	if vf.Kind() != reflect.Slice && vf.Kind() != reflect.Array {
		return []interface{}{v}
	}
	var r []interface{}
	for i := 0; i < vf.Len(); i++ {
		r = append(r, vf.Index(i).Interface())
	}
	return r
}

func mergeArray(i ...interface{}) []interface{} {
	var r []interface{}
	for _, v := range i {
		if v != nil {
			r = append(r, parseArray(v)...)
		}
	}
	return r
}

func build(query bson.M, node *Node) {
	//fmt.Printf("build query:%v,node:%+v\n", query, node)
	if _, ok := query[node.k]; !ok {
		if node.t == QueryOperationPrefix {
			query[node.k] = node.v[0]
		} else {
			m := bson.M{}
			if isArray(node) {
				m[node.t] = node.v
			} else {
				m[node.t] = node.v[0]
			}
			query[node.k] = m
		}
		return
	}

	var ok bool
	var oldVal bson.M
	if oldVal, ok = query[node.k].(bson.M); !ok {
		oldVal = bson.M{}
		oldVal["$in"] = []interface{}{query[node.k]}
		query[node.k] = oldVal
	}

	if isArray(node) {
		oldVal[node.t] = mergeArray(oldVal[node.t], node.v)
	} else {
		oldVal[node.t] = node.v[0]
	}
}
