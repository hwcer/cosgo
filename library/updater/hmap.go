package updater

import (
	"github.com/hwcer/cosgo/library/logger"
	"github.com/hwcer/cosgo/storage/mongo"
)

/*
HMap 适用于无限叠加道具
*/

//has 32 +k 32
type HMap struct {
	*Base
	model     modelTable
	binder    binder //绑定的数据源
	dataset   *Dataset
	bulkWrite *mongo.BulkWrite
}

func NewHMap(bag, sort int32, model modelTable, updater *Updater) *HMap {
	i := &HMap{Base: NewBase(bag, sort, updater), model: model, dataset: &Dataset{}}
	i.release()
	return i
}

func (this *HMap) release() {
	this.bulkWrite = nil
	this.Base.release()
	this.dataset.release()
}

//Val  拥有道具总数量
func (this *HMap) Val(id int32) (r int64) {
	dataset := this.Dataset()
	return dataset.Val(id)
}

//Get 根据道具对象
func (this *HMap) Get(id interface{}) (r itemHMap) {
	dataset := this.Dataset()
	return dataset.Get(id)
}

func (this *HMap) Add(k int32, v int64) {
	this.act(ActTypeAdd, k, v)
}

func (this *HMap) Sub(k int32, v int64) {
	this.act(ActTypeSub, k, v)
}

//Set id= iid||oid ,v=map[string]interface{}
func (this *HMap) Set(id interface{}, v interface{}) {
	val, ok := v.(map[string]interface{})
	if !ok {
		logger.Error("HMap set v error")
		return
	}

	iid, oid, err := this.parseId(id)
	if err != nil {
		logger.Error("%v", err)
		return
	}
	act := &Cache{ID: oid, Id: iid, T: ActTypeSet, K: "*", V: val, B: this.Base.Bag()}
	this.doAct(act)
}

func (this *HMap) act(t ActType, k int32, v int64) {
	oid, _ := this.model.ObjectId(this.updater.uid, k)
	if oid == "" {
		return
	}
	act := &Cache{ID: oid, Id: k, T: t, K: "val", V: v, B: this.Base.Bag()}
	//logger.Debug("HMap act:%+v", act)
	this.doAct(act)
}

func (this *HMap) doAct(act *Cache) {
	if act.T != ActTypeDel {
		this.Keys(act.ID)
	}
	this.Base.Act(act)
	if this.Base.verify {
		this.Parse(act)
	}
}

func (this *HMap) Data() error {
	query := this.Base.fields.Query()
	if query == nil {
		return nil
	}
	rows, err := this.model.UGet(this.updater.uid, query)
	if err != nil {
		return err
	}
	for _, d := range rows {
		this.dataset.Set(d.(itemHMap))
	}
	this.Base.fields.reset()
	return nil
}

func (this *HMap) Verify() (err error) {
	this.Base.verify = true
	if len(this.Base.acts) == 0 {
		return nil
	}
	for _, act := range this.Base.acts {
		if this.updater.subVerify && act.T == ActTypeSub {
			av, _ := act.V.(int64)
			data := this.Get(act.ID)
			if data == nil {
				return ErrItemNotEnough(act.Id, av, 0)
			}
			dv := data.GetVal()
			if dv < av {
				return ErrItemNotEnough(act.Id, av, dv)
			}
		}
		this.Parse(act)
	}
	return nil
}

func (this *HMap) Save() (ret []*Cache, err error) {
	ret = this.Base.acts
	this.Base.acts = nil
	this.Base.verify = false
	if this.bulkWrite != nil {
		_, err = this.bulkWrite.Save()
	}
	if err != nil {
		ret = nil
	}
	this.bulkWrite = nil
	return
}

func (this *HMap) parseId(id interface{}) (iid int32, oid string, err error) {
	switch id.(type) {
	case string:
		oid = id.(string)
		iid, err = this.model.Parse(oid)
	case int, int32, int64:
		iid = int32(ParseInt(iid))
		oid, err = this.model.ObjectId(this.updater.uid, iid)
	}
	return
}

//Bind 绑定一个hmap,table作为数据源
func (this *HMap) Bind(p binder) {
	this.binder = p
}

func (this *HMap) Keys(keys ...interface{}) {
	if this.binder != nil {
		this.binder.Keys(keys...)
	} else {
		this.Base.Keys(keys...)
	}
}
func (this *HMap) Fields(keys ...string) {
	if this.binder != nil {
		this.binder.Fields(keys...)
	} else {
		this.Base.Fields(keys...)
	}
}

func (this *HMap) Dataset() *Dataset {
	if this.binder != nil {
		return this.binder.Dataset()
	} else {
		return this.dataset
	}
}

func (this *HMap) BulkWrite() *mongo.BulkWrite {
	if this.binder != nil {
		return this.binder.BulkWrite()
	}
	if this.bulkWrite == nil {
		this.bulkWrite = this.model.BulkWrite()
	}
	return this.bulkWrite
}
