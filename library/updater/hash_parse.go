package updater

import (
	"errors"
)

var hashParseHandle = make(map[ActType]func(*Hash, *Cache) error)

func init() {
	hashParseHandle[ActTypeAdd] = hashHandleAdd
	hashParseHandle[ActTypeSet] = hashHandleSet
	hashParseHandle[ActTypeSub] = hashHandleSub
}

func (h *Hash) Parse(act *Cache) error {
	if f, ok := hashParseHandle[act.T]; ok {
		return f(h, act)
	}
	return errors.New("hash_act_parser not exist")
}

func hashHandleAdd(h *Hash, act *Cache) (err error) {
	v := act.V.(int64)
	act.R = h.dataset.Add(act.K, v)
	h.update.Inc(act.K, v)
	return
}

func hashHandleSub(h *Hash, act *Cache) (err error) {
	v := act.V.(int64)
	act.R = h.dataset.Sub(act.K, v)
	h.update.Inc(act.K, -v)
	return

}

func hashHandleSet(h *Hash, act *Cache) (err error) {
	act.R = h.dataset.Set(act.K, act.V)
	h.update.Set(act.K, act.V)
	return
}
