package schema

import "sync"
//New 新封装schema store namer
func New(opts ... interface{}) (i *Store) {
	var (
		store *sync.Map
		namer Namer
	)
	for _,v:=range opts{
		if s,ok := v.(*sync.Map);ok{
			store = s
		}else if n,ok2 := v.(Namer);ok2{
			namer = n
		}
	}
	if store == nil{
		store = &sync.Map{}
	}
	if namer == nil{
		namer = &NamingStrategy{}
	}
	i = &Store{store: store,namer: namer}
	return i
}

type Store struct {
	store *sync.Map
	namer Namer
}

func (s *Store) Parse(dest interface{}) (*Schema, error) {
	return Parse(dest, s.store, s.namer)
}

func (s *Store) ParseWithSpecialTableName(dest interface{}, specialTableName string) (*Schema, error) {
	return ParseWithSpecialTableName(dest, s.store, s.namer, specialTableName)
}
