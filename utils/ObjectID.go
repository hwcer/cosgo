package utils

import (
	"errors"
	"gopkg.in/mgo.v2/bson"
)

func ObjectID(s ...string) (bson.ObjectId, error) {
	if len(s) == 0 {
		return bson.NewObjectId(), nil
	} else if bson.IsObjectIdHex(s[0]) {
		return bson.ObjectIdHex(s[0]), nil
	} else {
		return "", errors.New("ObjectId args is invalid")
	}
}
