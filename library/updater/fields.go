package updater

import "github.com/hwcer/cosgo/storage/mongo"

type fields struct {
	fields  []interface{}
	history map[interface{}]bool
}

func NewFields() *fields {
	return &fields{
		history: make(map[interface{}]bool),
	}
}

func (s *fields) Keys(keys ...interface{}) {
	var arr []interface{}
	for _, k := range keys {
		if !s.Has(k) {
			arr = append(arr, k)
		}
	}
	s.fields = append(s.fields, arr...)
}

func (s *fields) Has(k interface{}) bool {
	if _, ok := s.history[k]; ok {
		return true
	}
	for _, f := range s.fields {
		if f == k {
			return true
		}
	}
	return false
}

func (s *fields) reset() {
	for _, k := range s.fields {
		s.history[k] = true
	}
	s.fields = nil
}

func (s *fields) release() {
	s.fields = nil
	s.history = make(map[interface{}]bool)
}

func (s *fields) String() (r []string) {
	for _, v := range s.fields {
		if s, ok := v.(string); ok {
			r = append(r, s)
		}
	}
	return
}

func (s *fields) Query() mongo.Query {
	if len(s.fields) == 0 {
		return nil
	}
	var oid []string
	var iid []interface{}
	for _, v := range s.fields {
		switch v.(type) {
		case string:
			oid = append(oid, v.(string))
		case int, int32, int64, uint, uint32, uint64:
			iid = append(iid, v)
		}
	}
	query := mongo.Query{}
	if len(oid) > 0 && len(iid) > 0 {
		q1 := mongo.Query{}
		q1.PrimaryString(oid...)
		q2 := mongo.Query{}
		q2.In(ItemNameId, iid...)
		query.OR(q1, q2)
	} else if len(oid) > 0 {
		query.PrimaryString(oid...)
	} else if len(iid) > 0 {
		query.In(ItemNameId, iid...)
	}
	return query
}
