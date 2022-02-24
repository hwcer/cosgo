package etcdv3

import (
	"context"
	"errors"
	"fmt"
	"github.com/hwcer/cosgo/logger"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/rpcxio/libkv"
	"github.com/rpcxio/libkv/store"
)

func init() {
	Register()
}

// EtcdV3RegisterPlugin implements etcdv3 registry.
type EtcdV3RegisterPlugin struct {
	// service address, for example, tcp@127.0.0.1:8972, quic@127.0.0.1:1234
	ServiceAddress string
	// etcdv3 addresses
	EtcdServers []string
	// base path for rpcx server, for example com/example/rpcx
	BasePath string
	Metrics  metrics.Registry
	// Registered services
	Services       []string
	metasLock      sync.RWMutex
	metas          map[string]string
	UpdateInterval time.Duration

	Options *store.Config
	kv      store.Store

	dying chan struct{}
	done  chan struct{}
}

// Start starts to connect etcdv3 cluster
func (p *EtcdV3RegisterPlugin) Start() error {
	if p.done == nil {
		p.done = make(chan struct{})
	}
	if p.dying == nil {
		p.dying = make(chan struct{})
	}

	if p.kv == nil {
		kv, err := libkv.NewStore(ETCDV3, p.EtcdServers, p.Options)
		if err != nil {
			logger.Error("cannot create etcdv3 registry: %v", err)
			return err
		}
		p.kv = kv
	}

	err := p.kv.Put(p.BasePath, []byte("rpcx_path"), &store.WriteOptions{IsDir: true})
	if err != nil && !strings.Contains(err.Error(), "Not a file") {
		return err
	}

	if p.UpdateInterval > 0 {
		ticker := time.NewTicker(p.UpdateInterval)
		go func() {
			defer p.kv.Close()
			// refresh service TTL
			for {
				select {
				case <-p.dying:
					close(p.done)
					return
				case <-ticker.C:
					extra := make(map[string]string)
					if p.Metrics != nil {
						extra["calls"] = fmt.Sprintf("%.2f", metrics.GetOrRegisterMeter("calls", p.Metrics).RateMean())
						extra["connections"] = fmt.Sprintf("%.2f", metrics.GetOrRegisterMeter("connections", p.Metrics).RateMean())
					}
					//set this same metrics for all services at this server
					for _, name := range p.Services {
						nodePath := fmt.Sprintf("%s/%s/%s", p.BasePath, name, p.ServiceAddress)
						kvPair, err := p.kv.Get(nodePath)
						if err != nil {
							p.metasLock.RLock()
							meta := p.metas[name]
							p.metasLock.RUnlock()

							err = p.kv.Put(nodePath, []byte(meta), nil)
							if err != nil {
								logger.Error("cannot re-create etcdv3 path %s: %v", nodePath, err)
							}

						} else {
							v, _ := url.ParseQuery(string(kvPair.Value))
							for key, value := range extra {
								v.Set(key, value)
							}
							p.kv.Put(nodePath, []byte(v.Encode()), nil)
						}
					}
				}
			}
		}()
	}

	return nil
}

// Stop unregister all services.
func (p *EtcdV3RegisterPlugin) Stop() error {
	if p.kv == nil {
		kv, err := libkv.NewStore(ETCDV3, p.EtcdServers, p.Options)
		if err != nil {
			return err
		}
		p.kv = kv
	}

	for _, name := range p.Services {
		nodePath := fmt.Sprintf("%s/%s/%s", p.BasePath, name, p.ServiceAddress)
		exist, err := p.kv.Exists(nodePath)
		if err != nil {
			logger.Error("cannot delete path %s: %v", nodePath, err)
			continue
		}
		if exist {
			p.kv.Delete(nodePath)
		}
	}

	close(p.dying)
	<-p.done
	return nil
}

// HandleConnAccept handles connections from clients
func (p *EtcdV3RegisterPlugin) HandleConnAccept(conn net.Conn) (net.Conn, bool) {
	if p.Metrics != nil {
		metrics.GetOrRegisterMeter("connections", p.Metrics).Mark(1)
	}
	return conn, true
}

// PreCall handles rpc call from clients
func (p *EtcdV3RegisterPlugin) PreCall(_ context.Context, _, _ string, args interface{}) (interface{}, error) {
	if p.Metrics != nil {
		metrics.GetOrRegisterMeter("calls", p.Metrics).Mark(1)
	}
	return args, nil
}

// Register handles registering event.
// this service is registered at BASE/serviceName/thisIpAddress node
func (p *EtcdV3RegisterPlugin) Register(name string, rcvr interface{}, metadata string) (err error) {
	if strings.TrimSpace(name) == "" {
		err = errors.New("Watch service `name` can't be empty")
		return
	}

	if p.kv == nil {
		Register()
		kv, err := libkv.NewStore(ETCDV3, p.EtcdServers, nil)
		if err != nil {
			return err
		}
		p.kv = kv
	}

	err = p.kv.Put(p.BasePath, []byte("rpcx_path"), &store.WriteOptions{IsDir: true})
	if err != nil && !strings.Contains(err.Error(), "Not a file") {
		return err
	}

	nodePath := fmt.Sprintf("%s/%s", p.BasePath, name)
	err = p.kv.Put(nodePath, []byte(name), &store.WriteOptions{IsDir: true})
	if err != nil && !strings.Contains(err.Error(), "Not a file") {
		return err
	}

	nodePath = fmt.Sprintf("%s/%s/%s", p.BasePath, name, p.ServiceAddress)
	err = p.kv.Put(nodePath, []byte(metadata), &store.WriteOptions{TTL: p.UpdateInterval * 2})
	if err != nil {
		return err
	}

	p.Services = append(p.Services, name)

	p.metasLock.Lock()
	if p.metas == nil {
		p.metas = make(map[string]string)
	}
	p.metas[name] = metadata
	p.metasLock.Unlock()
	return
}

func (p *EtcdV3RegisterPlugin) RegisterFunction(serviceName, fname string, fn interface{}, metadata string) error {
	return p.Register(serviceName, fn, metadata)
}

func (p *EtcdV3RegisterPlugin) Unregister(name string) (err error) {
	if len(p.Services) == 0 {
		return nil
	}

	if strings.TrimSpace(name) == "" {
		err = errors.New("Watch service `name` can't be empty")
		return
	}

	if p.kv == nil {
		Register()
		kv, err := libkv.NewStore(ETCDV3, p.EtcdServers, nil)
		if err != nil {
			return err
		}
		p.kv = kv
	}

	err = p.kv.Put(p.BasePath, []byte("rpcx_path"), &store.WriteOptions{IsDir: true})
	if err != nil && !strings.Contains(err.Error(), "Not a file") {
		return err
	}

	nodePath := fmt.Sprintf("%s/%s", p.BasePath, name)
	err = p.kv.Put(nodePath, []byte(name), &store.WriteOptions{IsDir: true})
	if err != nil && !strings.Contains(err.Error(), "Not a file") {
		return err
	}

	nodePath = fmt.Sprintf("%s/%s/%s", p.BasePath, name, p.ServiceAddress)

	err = p.kv.Delete(nodePath)
	if err != nil {
		return err
	}

	if len(p.Services) > 0 {
		var services = make([]string, 0, len(p.Services)-1)
		for _, s := range p.Services {
			if s != name {
				services = append(services, s)
			}
		}
		p.Services = services
	}

	p.metasLock.Lock()
	if p.metas == nil {
		p.metas = make(map[string]string)
	}
	delete(p.metas, name)
	p.metasLock.Unlock()
	return
}
