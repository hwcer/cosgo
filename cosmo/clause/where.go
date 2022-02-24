package clause

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
	"strings"
)

const sqlConditionSplit = " " //SQL语法分隔符

//Where 构造查询条件
//支持 =,>,<,>=,<=,<>,!=
//支持使用OR,AND,NOT,NOR连接多个条件，OR,AND,NOT,NOR一次只能拼接一种
var whereComplexMap = make(map[string]string)
var whereConditionArr = []string{"NIN", "IN", "!=", "<>", ">=", "<=", ">", "<", "="}
var whereConditionSql = make(map[string]string)
var whereConditionMongo = map[string]string{
	"=":   "",
	"!=":  "nin",
	"<>":  "nin",
	">=":  "gte",
	"<=":  "lte",
	">":   "gt",
	"<":   "lt",
	"IN":  "in",
	"NIN": "nin",
}

func init() {
	for _, k := range complexCondition {
		pair := []string{"", strings.ToUpper(k), ""}
		whereComplexMap[k] = strings.Join(pair, sqlConditionSplit)
	}

	for _, k := range whereConditionArr {
		pair := []string{"", strings.ToUpper(k), ""}
		whereConditionSql[k] = strings.Join(pair, sqlConditionSplit)
	}

}

func IsQueryFormat(s string) bool {
	for _, k := range whereConditionArr {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}

func (q *Query) fromMap(i reflect.Value) (query string, args []interface{}) {
	var pairs []string
	for _, key := range i.MapKeys() {
		pairs = append(pairs, fmt.Sprintf("%v = ?", key.Interface().(string)))
		args = append(args, i.MapIndex(key).Interface())
	}
	query = strings.Join(pairs, whereComplexMap[QueryOperationAND])
	return
}

//TODO
func (q *Query) fromStruct(i interface{}) (query string, args []interface{}) {

	return
}

func (q *Query) Where(format interface{}, cons ...interface{}) {
	var args []interface{}
	var query string
	vof := reflect.Indirect(reflect.ValueOf(format))

	switch vof.Kind() {
	case reflect.String:
		args = cons
		query = format.(string)
	case reflect.Map:
		query, args = q.fromMap(vof)
	case reflect.Struct:
		query, args = q.fromStruct(vof)
	default:
		args = cons
	}
	//query 非查询语句时，使用主键匹配
	if !IsQueryFormat(query) {
		query = MongoPrimaryName + " = ?"
		args = append([]interface{}{format}, args...)
	}

	var arr []string
	var whereType string
	for _, k := range complexCondition {
		if strings.Contains(query, whereComplexMap[k]) {
			whereType = k
			arr = strings.Split(query, whereComplexMap[k])
			break
		}
	}
	if len(arr) == 0 {
		arr = append(arr, query)
		whereType = QueryOperationAND
	}

	var nodes []*Node
	var argIndex int = 0

	for _, pair := range arr {
		var v interface{}
		if strings.Contains(pair, "?") && argIndex < len(args) {
			v = args[argIndex]
			argIndex += 1
		}
		for _, w := range whereConditionArr {
			//fmt.Printf("Where pair: %v,w:%v, sql:%v \n", pair, w, whereConditionSql[w])
			if strings.Contains(pair, whereConditionSql[w]) {
				if node := parseWherePair(pair, w, v); node != nil {
					nodes = append(nodes, node)
				}
				break
			}
		}
	}
	if whereType == QueryOperationAND {
		for _, node := range nodes {
			if node.k == MongoPrimaryName && node.t == QueryOperationPrefix {
				q.Primary(node.v)
			} else {
				q.where = append(q.where, node)
			}
		}
		return
	}
	var con []interface{}
	for _, node := range nodes {
		c := make(bson.M)
		build(c, node)
		con = append(con, c)
	}
	q.match(whereType, con)
}

func parseWherePair(pair string, w string, v interface{}) *Node {
	//fmt.Printf("parseWherePair %v---%v ---%v\n", pair, w, v)
	arr := strings.Split(pair, w)
	//fmt.Printf("parseWherePair ARR: %v \n", arr)
	if len(arr) != 2 {
		return nil
	}
	node := &Node{}
	node.t = QueryOperationPrefix + whereConditionMongo[w]
	node.k = strings.Trim(arr[0], sqlConditionSplit)

	var r interface{}
	r = strings.Trim(arr[1], sqlConditionSplit)
	if r == "?" {
		r = v
	}
	node.v = parseArray(r)
	if node.t == QueryOperationPrefix && len(node.v) > 1 {
		node.t = "$in"
	}
	//fmt.Printf("parseWherePair node: %+v \n", node)
	return node
}
