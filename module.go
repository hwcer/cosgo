package cosgo

type Module interface {
	Id() string
	Init() error
	Start() error
	Close() error
}

type ModuleReload interface {
	Reload() error
}

//
//func NewModule(id string) *ModuleDefault {
//	return &ModuleDefault{id: id}
//}
//
//type ModuleDefault struct {
//	id string
//}
//
//func (m *ModuleDefault) Id() string {
//	return m.id
//}
//
//func (m *ModuleDefault) Init() error {
//	return nil
//}
//
//func (m *ModuleDefault) Start() error {
//	return nil
//}
//
//func (m *ModuleDefault) Close() error {
//	return nil
//}
//
//func (m *ModuleDefault) Reload() error {
//	return nil
//}
