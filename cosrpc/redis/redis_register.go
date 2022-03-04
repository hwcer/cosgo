package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/hwcer/cosgo/library/logger"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/rpcxio/libkv"
	"github.com/rpcxio/libkv/store"
	"github.com/smallnest/rpcx/log"
)

func init() {
	Register()
}

type rpcxServerMetadataFunc func(servicesName string, metadata string) string

// RedisRegisterPlugin implements redis registry.
type RedisRegisterPlugin struct {
	// service address, for example, tcp@127.0.0.1:8972, quic@127.0.0.1:1234
	ServiceAddress string
	// redis addresses
	RedisServers []string
	// base path for rpcx server, for example com/example/rpcx
	BasePath string
	Metrics  metrics.Registry
	Metadata interface{}
	// Registered services
	Services       []string
	metasLock      sync.RWMutex
	metas          map[string]string
	UpdateInterval time.Duration
	Options        *store.Config
	kv             store.Store
	done           chan struct{}
	dying          chan struct{}
}

func (p *RedisRegisterPlugin) writeOptions(isDir bool) *store.WriteOptions {
	o := &store.WriteOptions{IsDir: isDir}
	if !isDir {
		o.TTL = p.UpdateInterval + 5*time.Second
	}
	return o
}

func (p *RedisRegisterPlugin) GetMetadata(servicesName string, metadata string) string {
	if p.Metadata == nil {
		return metadata
	}
	switch p.Metadata.(type) {
	case string:
		return p.Metadata.(string) + "&" + metadata
	case rpcxServerMetadataFunc:
		return p.Metadata.(rpcxServerMetadataFunc)(servicesName, metadata)
	}
	return metadata
}

// Start starts to connect redis cluster
func (p *RedisRegisterPlugin) Start() error {
	if p.dying != nil {
		return nil
	}
	p.done = make(chan struct{})
	p.dying = make(chan struct{})

	if p.kv == nil {
		kv, err := libkv.NewStore(RPCREDIS, p.RedisServers, p.Options)
		if err != nil {
			log.Errorf("cannot create redis registry: %v", err)
			return err
		}
		p.kv = kv
	}

	err := p.kv.Put(p.BasePath, []byte("rpcx_path"), p.writeOptions(true))
	if err != nil && !strings.Contains(err.Error(), "Not a file") {
		log.Errorf("cannot create redis path %s: %v", p.BasePath, err)
		return err
	}
	if p.UpdateInterval > 0 {
		go p.doUpdateInterval()
	} else {
		close(p.done)
	}

	return nil
}

// Stop unregister all services.
func (p *RedisRegisterPlugin) Stop() error {
	if p.kv == nil {
		kv, err := libkv.NewStore(RPCREDIS, p.RedisServers, p.Options)
		if err != nil {
			log.Errorf("cannot create redis registry: %v", err)
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
func (p *RedisRegisterPlugin) HandleConnAccept(conn net.Conn) (net.Conn, bool) {
	if p.Metrics != nil {
		metrics.GetOrRegisterMeter("connections", p.Metrics).Mark(1)
	}
	return conn, true
}

// PreCall handles rpc call from clients
func (p *RedisRegisterPlugin) PreCall(_ context.Context, _, _ string, args interface{}) (interface{}, error) {
	if p.Metrics != nil {
		metrics.GetOrRegisterMeter("calls", p.Metrics).Mark(1)
	}
	return args, nil
}

// Register handles registering event.
// this service is registered at BASE/serviceName/thisIpAddress node
func (p *RedisRegisterPlugin) Register(name string, rcvr interface{}, metadata string) (err error) {
	for _, k := range p.Services {
		if k == name {
			return
		}
	}
	if strings.TrimSpace(name) == "" {
		err = errors.New("Watch service `name` can't be empty")
		return
	}

	if p.kv == nil {
		Register()
		kv, err := libkv.NewStore(RPCREDIS, p.RedisServers, p.Options)
		if err != nil {
			log.Errorf("cannot create redis registry: %v", err)
			return err
		}
		p.kv = kv
	}
	metadata = p.GetMetadata(name, metadata)
	err = p.kv.Put(p.BasePath, []byte("rpcx_path"), p.writeOptions(true))
	if err != nil && !strings.Contains(err.Error(), "Not a file") {
		log.Errorf("cannot create redis path %s: %v", p.BasePath, err)
		return err
	}

	nodePath := fmt.Sprintf("%s/%s", p.BasePath, name)
	err = p.kv.Put(nodePath, []byte(name), p.writeOptions(true))
	if err != nil && !strings.Contains(err.Error(), "Not a file") {
		log.Errorf("cannot create redis path %s: %v", nodePath, err)
		return err
	}

	nodePath = fmt.Sprintf("%s/%s/%s", p.BasePath, name, p.ServiceAddress)
	err = p.kv.Put(nodePath, []byte(metadata), p.writeOptions(false))
	if err != nil {
		log.Errorf("cannot create redis path %s: %v", nodePath, err)
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
func (p *RedisRegisterPlugin) RegisterFunction(serviceName, name string, fn interface{}, metadata string) (err error) {
	return p.Register(serviceName, fn, metadata)
}

func (p *RedisRegisterPlugin) Unregister(name string) (err error) {
	if len(p.Services) == 0 {
		return nil
	}

	if strings.TrimSpace(name) == "" {
		err = errors.New("Watch service `name` can't be empty")
		return
	}

	if p.kv == nil {
		Register()
		kv, err := libkv.NewStore(RPCREDIS, p.RedisServers, p.Options)
		if err != nil {
			log.Errorf("cannot create redis registry: %v", err)
			return err
		}
		p.kv = kv
	}

	err = p.kv.Put(p.BasePath, []byte("rpcx_path"), p.writeOptions(true))
	if err != nil && !strings.Contains(err.Error(), "Not a file") {
		log.Errorf("cannot create redis path %s: %v", p.BasePath, err)
		return err
	}

	nodePath := fmt.Sprintf("%s/%s", p.BasePath, name)
	err = p.kv.Put(nodePath, []byte(name), p.writeOptions(true))
	if err != nil && !strings.Contains(err.Error(), "Not a file") {
		log.Errorf("cannot create redis path %s: %v", nodePath, err)
		return err
	}

	nodePath = fmt.Sprintf("%s/%s/%s", p.BasePath, name, p.ServiceAddress)

	err = p.kv.Delete(nodePath)
	if err != nil {
		log.Errorf("cannot create consul path %s: %v", nodePath, err)
		return err
	}

	var services = make([]string, 0, len(p.Services)-1)
	for _, s := range p.Services {
		if s != name {
			services = append(services, s)
		}
	}
	p.Services = services

	p.metasLock.Lock()
	if p.metas == nil {
		p.metas = make(map[string]string)
	}
	delete(p.metas, name)
	p.metasLock.Unlock()
	return
}

func (p *RedisRegisterPlugin) doUpdateInterval() {
	defer close(p.done)
	defer p.kv.Close()
	ticker := time.NewTicker(p.UpdateInterval)
	defer ticker.Stop()
	// refresh service TTL
	for {
		select {
		case <-p.dying:
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
					meta = p.GetMetadata(name, meta)
					err = p.kv.Put(nodePath, []byte(meta), p.writeOptions(false))
					if err != nil {
						logger.Error("cannot re-create redis path %s: %v", nodePath, err)
					}
				} else {
					v, _ := url.ParseQuery(string(kvPair.Value))
					for key, value := range extra {
						v.Set(key, value)
					}
					p.kv.Put(nodePath, []byte(v.Encode()), p.writeOptions(false))
				}
			}
		}
	}
}
