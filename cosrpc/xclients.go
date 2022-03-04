package cosrpc

import (
	"context"
	"fmt"
	"github.com/hwcer/cosgo/library/logger"
	rpcx "github.com/smallnest/rpcx/client"
	"sync/atomic"
)

func NewXClient() *XClient {
	return &XClient{
		clients: make(map[string]*Options),
	}
}

type Options struct {
	client    rpcx.XClient
	selector  rpcx.Selector
	selectMod rpcx.SelectMode
	discovery rpcx.ServiceDiscovery
	Option    rpcx.Option
	FailMode  rpcx.FailMode
}

type XClient struct {
	start     int32
	clients   map[string]*Options
	discovery rpcx.ServiceDiscovery
}

//AddServicePath 观察服务器信息
func (this *XClient) AddServicePath(servicePath string, selector interface{}) (opts *Options) {
	if atomic.LoadInt32(&this.start) > 0 {
		logger.Error("XClient已经启动无法再添加Service")
	}
	opts = &Options{}
	this.clients[servicePath] = opts
	switch selector.(type) {
	case rpcx.Selector:
		opts.selector = selector.(rpcx.Selector)
		opts.selectMod = rpcx.SelectByUser
	case rpcx.SelectMode:
		opts.selectMod = selector.(rpcx.SelectMode)
	default:
		logger.Fatal("XClient AddServicePath arg(Selector) type error:%v", selector)
	}
	opts.Option = rpcx.DefaultOption
	opts.FailMode = rpcx.Failtry
	return
}

func (this *XClient) Start(discovery rpcx.ServiceDiscovery) (err error) {
	if !atomic.CompareAndSwapInt32(&this.start, 0, 1) {
		return
	}
	if len(this.clients) == 0 {
		return
	}
	this.discovery = discovery
	for k, _ := range this.clients {
		if err = this.create(k); err != nil {
			return
		}
	}
	return
}

func (this *XClient) Close() (errs []error) {
	var err error
	for _, c := range this.clients {
		if err = c.client.Close(); err != nil {
			errs = append(errs, err)
		}
		c.discovery.Close()
	}
	this.discovery.Close()
	return nil
}

func (this *XClient) create(servicePath string) (err error) {
	c := this.clients[servicePath]
	c.discovery = this.discovery
	//if c.discovery, err = this.discovery.Clone(servicePath); err != nil {
	//	return
	//}
	c.client = rpcx.NewXClient(servicePath, c.FailMode, c.selectMod, c.discovery, c.Option)
	if c.selectMod == rpcx.SelectByUser {
		c.client.SetSelector(c.selector)
	}
	return
}
func (this *XClient) Has(servicePath string) bool {
	if _, ok := this.clients[servicePath]; ok {
		return true
	} else {
		return false
	}
}

func (this *XClient) Size() int {
	return len(this.clients)
}

func (this *XClient) Client(servicePath string) rpcx.XClient {
	if v, ok := this.clients[servicePath]; ok {
		return v.client
	} else {
		return nil
	}
}

func (this *XClient) Call(servicePath, serviceMethod string, args, reply interface{}) (err error) {
	if c := this.Client(servicePath); c != nil {
		return c.Call(context.Background(), serviceMethod, args, reply)
	} else {
		return fmt.Errorf("服务不存在")
	}
}

func (this *XClient) Broadcast(servicePath, serviceMethod string, args, reply interface{}) (err error) {
	if c := this.Client(servicePath); c != nil {
		return c.Broadcast(context.Background(), serviceMethod, args, reply)
	} else {
		return fmt.Errorf("服务不存在")
	}
}
