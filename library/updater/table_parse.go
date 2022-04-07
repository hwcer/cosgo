package updater

import (
	"fmt"
	"github.com/hwcer/cosgo/storage/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

var tableParseHandle = make(map[ActType]func(*Table, *Cache) error)

func init() {
	tableParseHandle[ActTypeNew] = tableHandleNew
	tableParseHandle[ActTypeAdd] = tableHandleAdd
	tableParseHandle[ActTypeDel] = tableHandleDel
	tableParseHandle[ActTypeSet] = tableHandleSet
}

func (this *Table) Parse(act *Cache) error {
	if f, ok := tableParseHandle[act.T]; ok {
		return f(this, act)
	}
	return fmt.Errorf("table_act_parse not exist:%v", act.T)
}

func tableHandleDel(t *Table, act *Cache) error {
	t.Dataset().Del(act.ID)
	bulkWrite := t.BulkWrite()
	query := mongo.NewPrimaryQuery(act.ID)
	bulkWrite.Delete(query, false)
	return nil
}

func tableHandleNew(t *Table, act *Cache) (err error) {
	bulkWrite := t.BulkWrite()
	id := act.Id
	if act.R, err = t.model.New(t.updater.uid, id, 1, act.B); err != nil {
		return
	}
	v := act.R.(itemHMap)
	act.ID = v.GetOId()
	t.Dataset().Set(v)
	act.T = ActTypeNew
	bulkWrite.Insert(act.R)
	return nil
}

//tableHandleAdd 必然是unique模式下才有
func tableHandleAdd(t *Table, act *Cache) (err error) {
	id := act.Id
	num := act.V.(int64)
	if len(t.Dataset().Indexes(id)) == 0 {
		if err = tableHandleNew(t, act); err != nil {
			return
		}
		num -= 1
	} else {
		act.T = ActTypeResolve
	}

	if num <= 0 {
		return
	}
	//自动分解
	m, _ := t.model.(resolve)
	if newId, newNum, ok := m.Resolve(id, int32(num)); ok {
		t.updater.Add(newId, newNum)
	}
	return
}

func tableHandleSet(t *Table, act *Cache) error {
	item := t.Get(act.ID)
	if item == nil {
		return fmt.Errorf("updater table set,item not exist:%v", act.ID)
	}
	bulkWrite := t.BulkWrite()
	act.R = act.V
	item.Set(act.V.(map[string]interface{}))
	query := mongo.NewPrimaryQuery(act.ID)
	v := bson.M{}
	v["$set"] = act.V
	update := mongo.Update(v)
	bulkWrite.Update(query, update)
	return nil
}
