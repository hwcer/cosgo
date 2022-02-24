package mongo

import (
	"testing"
)

type doc struct {
	Id    string `bson:"_id"`
	Name  string `bson:"name"`
	Value int    `bson:"value"`
	Flag  int    `bson:"flag"`
}

func TestNew(t *testing.T) {
	db, err := New("mongodb://127.0.0.1")
	if err != nil {
		t.Logf("err:%v", err)
		return
	}
	coll := db.Collection("test", "foo")

	_, err = coll.Insert(&doc{Id: ObjectId(), Name: "a", Value: 1, Flag: 1}, &doc{Id: ObjectId(), Name: "b", Value: 2, Flag: 1})
	if err != nil {
		t.Logf("err:%v", err)
		return
	}
	var results []*doc
	query := make(Query)
	query.Eq("flag", 1)
	update := make(Update)
	update.Inc("value", 2)
	update.Set("name", "newName")
	if _, err := coll.Update(query, update, true); err != nil {
		t.Logf("err:%v", err)
		return
	}

	page := NewPage(1, 5)
	page.Sort("Value", 1)
	page.Sort("Flag", 1)
	if _, err := coll.Page(query, page, &results); err != nil {
		t.Logf("err:%v", err)
		return
	}
	for _, v := range results {
		t.Logf("result:%+v", v)
	}

}
