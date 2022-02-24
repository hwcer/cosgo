package mongo

import "gopkg.in/mgo.v2/bson"

func ObjectId() string {
	return bson.NewObjectId().Hex()
}

type ArrString []string

func (a ArrString) Interface() (r []interface{}) {
	for _, v := range a {
		r = append(r, v)
	}
	return
}

type ArrInt []int

func (a ArrInt) Interface() (r []interface{}) {
	for _, v := range a {
		r = append(r, v)
	}
	return
}

type ArrInt32 []int32

func (a ArrInt32) Interface() (r []interface{}) {
	for _, v := range a {
		r = append(r, v)
	}
	return
}
