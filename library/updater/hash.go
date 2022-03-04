package updater

import (
	"github.com/hwcer/cosgo/library/logger"
	"github.com/hwcer/cosgo/storage/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type Hash struct {
	*Base
	model    modelHash
	update   mongo.Update
	dataset  Data
	objectId string
}

func NewHash(bag, sort int32, model modelHash, updater *Updater) *Hash {
	return &Hash{Base: NewBase(bag, sort, updater), model: model, dataset: make(Data)}
}

func (h *Hash) release() {
	h.update = nil
	h.dataset = make(Data)
	h.objectId = ""
	h.Base.release()
}

func (h *Hash) Get(id string) interface{} {
	return h.dataset.Get(id)
}

func (h *Hash) MGet(keys ...string) bson.M {
	ret := make(bson.M, len(keys))
	for _, k := range keys {
		ret[k] = h.dataset[k]
	}
	return ret
}

func (h *Hash) Val(k string) int64 {
	return h.dataset.GetInt(k)
}

func (h *Hash) Add(k string, v int64) {
	if v > 0 {
		h.Act(ActTypeAdd, k, v)
	}
}

func (h *Hash) Sub(k string, v int64) {
	if v > 0 {
		h.Act(ActTypeSub, k, v)
	}
}

func (h *Hash) Set(id interface{}, v interface{}) {
	k, ok := id.(string)
	if !ok {
		logger.Error("hash set id must string")
	}
	h.Act(ActTypeSet, k, v)
}

func (h *Hash) Act(t ActType, k string, v interface{}) {
	h.Keys(k)
	act := &Cache{
		ID: h.updater.uid,
		T:  t,
		K:  k,
		V:  v,
		B:  h.Bag(),
	}
	h.Base.Act(act)
	if h.Base.verify {
		h.Parse(act)
	}
}

func (h *Hash) Verify() error {
	h.Base.verify = true
	if len(h.Base.acts) == 0 {
		return nil
	}
	h.update = mongo.Update{}
	for _, act := range h.Base.acts {
		if h.updater.subVerify && act.T == ActTypeSub {
			dv := h.dataset.GetInt(act.K)
			v, _ := act.V.(int64)
			if v > dv {
				return ErrItemNotEnough(act.Id, v, dv)
			}
		}
		err := h.Parse(act)
		if err != nil {
			return err
		}
	}
	if im, ok := h.model.(modelSetOnInert); ok {
		iv := im.SetOnInert(h.updater.uid, h.updater.time)

		for k, v := range iv {
			h.update.SetOnInert(k, v)
		}
	}
	return nil
}

func (h *Hash) Save() (ret []*Cache, err error) {
	h.Base.verify = false
	acts := h.Base.acts
	h.Base.acts = nil

	if h.update == nil || len(acts) == 0 {
		return
	}
	var oid string
	if oid, err = h.ObjectId(); err != nil {
		return
	}
	err = h.model.USet(oid, h.update)
	if err == nil {
		ret = acts
	}
	return
}

func (h *Hash) Data() (err error) {
	keys := h.Base.fields.String()
	var oid string
	if oid, err = h.ObjectId(); err != nil {
		return
	}
	var data bson.M
	if data, err = h.model.UGet(oid, keys); err != nil {
		return
	}

	for k, v := range data {
		h.dataset.Set(k, v)
	}
	h.Base.fields.reset()
	return nil
}
func (h *Hash) ObjectId() (oid string, err error) {
	if h.objectId != "" {
		return h.objectId, nil
	}
	if oid, err = h.model.ObjectId(h.updater.uid, h.updater.time); err == nil {
		h.objectId = oid
	}
	return
}
