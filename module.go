package cosgo

type Module interface {
	ID() string
	Init() error
	Start() error
	Close() error
}

//type Module struct {
//	Id string
//}
//
//func (m *Module) ID() string {
//	return m.Id
//}
//
//func (m *Module) Init() error {
//	return nil
//}
//
//func (m *Module) Start() error {
//	return nil
//}
//
//func (m *Module) Close() error {
//	return nil
//}
