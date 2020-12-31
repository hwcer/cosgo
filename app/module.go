package app

import "sync"

type Module interface {
	ID() string
	Init() error
	Start(*sync.WaitGroup) error
	Close(*sync.WaitGroup) error
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

func (m *DefModule) Start(wg *sync.WaitGroup) error {
	wg.Add(1)
	return nil
}

func (m *DefModule) Close(wg *sync.WaitGroup) error {
	wg.Done()
	return nil
}
