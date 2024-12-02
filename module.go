package cosgo

type IModule interface {
	Id() string
	Init() error
	Start() error
	Close() error
	Reload() error
}

func NewModule(id string) *Module {
	return &Module{id: id}
}

type Module struct {
	id string
}

func (m *Module) Id() string {
	return m.id
}

func (m *Module) Init() error {
	return nil
}

func (m *Module) Start() error {
	return nil
}

func (m *Module) Close() error {
	return nil
}

func (m *Module) Reload() error {
	return nil
}
