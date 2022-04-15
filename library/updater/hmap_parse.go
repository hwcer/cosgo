package updater

import (
	"fmt"
	"github.com/hwcer/cosgo/storage/mongo"
)

type hmapParse func(*HMap, *Cache) error

var hmapParseHandle = make(map[ActType]hmapParse)

func init() {
	hmapParseHandle[ActTypeAdd] = hmapHandleAdd
	hmapParseHandle[ActTypeSub] = hmapHandleSub
	hmapParseHandle[ActTypeSet] = hmapHandleSet
}

func (this *HMap) Parse(act *Cache) (err error) {
	var f hmapParse
	var ok bool
	if f, ok = hmapParseHandle[act.T]; !ok {
		return fmt.Errorf("table_act_parse not exist:%v", act.T)
	}
	if d := this.Dataset().Indexes(act.Id); len(d) == 0 && act.T != ActTypeSub {
		return hmapHandleNew(this, act)
	} else {
		return f(this, act)
	}
}

func hmapHandleNew(h *HMap, act *Cache) (err error) {
	bulkWrite := h.BulkWrite()
	//溢出判断
	if act.T == ActTypeAdd {
		v := act.V.(int64)
		imax := Config.IMax(act.Id)
		if imax > 0 && v > imax {
			h.updater.overflow[act.Id] += v - imax
			act.V = imax
		}
	}

	var data interface{}
	if data, err = h.model.New(h.updater.uid, act.Id, 0, act.B); err != nil {
		return
	}
	r := data.(itemHMap)
	switch act.T {
	case ActTypeAdd:
		r.Add(act.V.(int64))
	//case ActTypeSub:
	//	r.Create(-act.Val.(int64))
	case ActTypeSet:
		r.Set(act.V.(map[string]interface{}))
	}
	h.Dataset().Set(r.Clone().(itemHMap))
	act.R = r
	act.T = ActTypeNew
	bulkWrite.Insert(r)
	return
}

func hmapHandleAdd(h *HMap, act *Cache) error {
	bulkWrite := h.BulkWrite()
	v := act.V.(int64)
	imax := Config.IMax(act.Id)
	data := h.Get(act.ID)
	rv := v + data.GetVal()
	if imax > 0 && data.GetVal() >= imax {
		h.updater.overflow[act.Id] += v
		act.V = 0
		act.T = ActTypeOverflow
		return nil
	} else if imax > 0 && rv > imax {
		ov := rv - imax
		h.updater.overflow[act.Id] += ov
		act.V = v - ov
	}

	act.R = data.Add(v)
	query := mongo.NewPrimaryQuery(act.ID)
	update := mongo.Update{}
	update.Inc(ItemNameVal, v)
	bulkWrite.Update(query, update)
	return nil
}

func hmapHandleSub(h *HMap, act *Cache) error {
	bulkWrite := h.BulkWrite()
	v := act.V.(int64)
	data := h.Get(act.ID)
	act.R = data.Sub(v)
	query := mongo.NewPrimaryQuery(act.ID)
	update := mongo.Update{}
	update.Inc(ItemNameVal, -v)
	bulkWrite.Update(query, update)
	return nil
}

func hmapHandleSet(h *HMap, act *Cache) error {
	bulkWrite := h.BulkWrite()
	act.R = act.V
	data := h.Get(act.ID)
	data.Set(act.V.(map[string]interface{}))
	query := mongo.NewPrimaryQuery(act.ID)
	update := mongo.Update(act.V.(map[string]interface{}))
	bulkWrite.Update(query, update)
	return nil
}
