package updater

import "github.com/hwcer/cosgo/storage/mongo"

type Setter map[string]bool

func NewSetter() Setter {
	return make(Setter)
}

func (s Setter) Add(keys ...string) Setter {
	for _, k := range keys {
		s[k] = true
	}
	return s
}

func (s Setter) Del(k string) Setter {
	delete(s, k)
	return s
}

func (s Setter) Has(k string) bool {
	return s[k]
}

func (s *Setter) Clear() {
	*s = make(Setter)
}

func (s Setter) Size() int {
	return len(s)
}

func (s Setter) ToString() (ret []string) {
	for k, _ := range s {
		ret = append(ret, k)
	}
	return
}

func (s Setter) ToPrimaryQuery() mongo.Query {
	var ids []interface{}
	for k, _ := range s {
		ids = append(ids, k)
	}
	return mongo.NewPrimaryQuery(ids...)
}
