package app

type Module interface {
	ID() string
	Init() error
	Start() error
	Close() error
}

type DefModule struct {
	Id string
}

func (m *DefModule) ID() string {
	return m.Id
}

func (m *DefModule) Init() error {
	return nil
}

func (m *DefModule) Start() error {
	return nil
}

func (m *DefModule) Close() error {
	return nil
}
