package main

import (
	"cosgo/cosnet"
	"sync"
)

func NewTcpMod(name string) *tcpMod {
	return &tcpMod{name: name}
}

type tcpMod struct {
	name   string
	antnet cosnet.Server
}

func (this *tcpMod) ID() string {
	return this.name
}

func (this *tcpMod) Load() (err error) {
	this.antnet, err = cosnet.New("tcp://0.0.0.0:3000", cosnet.MsgTypeMsg, &cosnet.DefMsgHandler{})
	return
}

func (this *tcpMod) Start(wgp *sync.WaitGroup) (err error) {
	wgp.Add(1)
	return
	//return this.antnet.Start()
}

func (this *tcpMod) Close(wgp *sync.WaitGroup) (err error) {
	//this.antnet.Close()
	wgp.Done()
	return
}
