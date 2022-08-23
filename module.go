package app

type Module interface {
	ID() string
	Init() error
	Start() error
	Close() error
}

type ModuleDefault struct {
	Id string
}

func (m *ModuleDefault) ID() string {
	return m.Id
}

func (m *ModuleDefault) Init() error {
	return nil
}

func (m *ModuleDefault) Start() error {
	return nil
}

func (m *ModuleDefault) Close() error {
	return nil
}
