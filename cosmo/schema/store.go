package schema

import "sync"

func New(store *sync.Map, namer Namer) (i *Store) {
	i = &Store{store: store}
	if namer != nil {
		i.namer = namer
	} else {
		i.namer = &NamingStrategy{}
	}
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
