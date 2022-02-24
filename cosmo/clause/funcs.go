package clause

import "go.mongodb.org/mongo-driver/bson"

//Multiple 使用主键判断是批量操作还是单个文档操作
func Multiple(query bson.M) bool {
	v, ok := query[MongoPrimaryName]
	if !ok {
		return true
	}
	switch v.(type) {
	case map[string]interface{}, bson.M:
		return true
	}
	return false
}
