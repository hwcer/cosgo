package app

import (
	"context"
	"fmt"
	"icefire/confer"
	"icefire/logger"
	"icefire/pb"
	"icefire/srpc"
	"sync"
)

func NewModBase(name string, id int) *ModBaseImpl {
	return &ModBaseImpl{
		Sid:      fmt.Sprintf("%v.%v", name, id),
		Idx:      id,
		Cfg:      nil,
		Srv:      nil,
		Wgp:      nil,
		Services: make(map[string]interface{}),
	}
}

type ModBaseImpl struct {
	Sid string
	Idx int
	Cfg confer.Confer
	Srv *srpc.SRpcServer
	Wgp *sync.WaitGroup

	Services map[string]interface{}
}

func (m *ModBaseImpl) ID() string {
	return m.Sid
}

func (m *ModBaseImpl) Register(name string, rcvr interface{}) {
	m.Services[name] = rcvr
}

func (m *ModBaseImpl) DoRegister() error {
	if m.Srv == nil {
		if len(m.Services) > 0 {
			logger.WARN("nil srv with some services")
		}
		return nil
	}

	for name, r := range m.Services {
		err := m.Srv.RegisterName(name, r, "")
		if err != nil {
			return err
		}
		logger.INFO("mod [%v] service [%v] register", m.ID(), name)
	}
	return nil
}

func (m *ModBaseImpl) Prepare(f FlagItf) error {
	return nil
}

func (m *ModBaseImpl) Init(c confer.Confer, s *srpc.SRpcServer) error {
	m.Cfg = c
	m.Srv = s

	err := m.DoRegister()
	if err != nil {
		return err
	}

	logger.INFO("mod [%v] init", m.ID())
	return nil
}

func (m *ModBaseImpl) Start(cx context.Context, wgp *sync.WaitGroup) error {
	m.Wgp = wgp
	m.Wgp.Add(1)

	logger.INFO("mod [%v] start", m.ID())
	return nil
}

func (m *ModBaseImpl) Stop() error {
	defer m.Wgp.Done()

	logger.INFO("mod [%v] stop", m.ID())
	return nil
}

func (m *ModBaseImpl) HandleAppCmd(cx context.Context, param *pb.SSAppCmd) error {
	logger.INFO("mod [%v] handle app cmd:%v", m.ID(), param.Cmd)
	return nil
}
