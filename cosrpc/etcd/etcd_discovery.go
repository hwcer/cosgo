package etcd

import (
	"github.com/nacos-group/nacos-sdk-go/common/logger"
	"strings"
	"sync"
	"time"

	"github.com/rpcxio/libkv"
	"github.com/rpcxio/libkv/store"
	"github.com/smallnest/rpcx/client"
)

func init() {
	Register()
}

// EtcdDiscovery is a etcd service discovery.
// It always returns the registered servers in etcd.
type EtcdDiscovery struct {
	basePath string
	kv       store.Store
	pairsMu  sync.RWMutex
	pairs    []*client.KVPair
	chans    []chan []*client.KVPair
	mu       sync.Mutex

	// -1 means it always retry to watch until zookeeper is ok, 0 means no retry.
	RetriesAfterWatchFailed int

	filter           client.ServiceDiscoveryFilter
	AllowKeyNotFound bool

	stopCh chan struct{}
}

// NewEtcdDiscovery returns a new EtcdDiscovery.
func NewEtcdDiscovery(basePath string, servicePath string, etcdAddr []string, allowKeyNotFound bool, options *store.Config) (client.ServiceDiscovery, error) {
	kv, err := libkv.NewStore(ETCD, etcdAddr, options)
	if err != nil {
		panic(err)
	}

	return NewEtcdDiscoveryStore(basePath+"/"+servicePath, kv, allowKeyNotFound)
}

// NewEtcdDiscoveryStore return a new EtcdDiscovery with specified store.
func NewEtcdDiscoveryStore(basePath string, kv store.Store, allowKeyNotFound bool) (client.ServiceDiscovery, error) {
	if len(basePath) > 1 && strings.HasSuffix(basePath, "/") {
		basePath = basePath[:len(basePath)-1]
	}

	d := &EtcdDiscovery{basePath: basePath, kv: kv}
	d.stopCh = make(chan struct{})

	ps, err := kv.List(basePath)
	if err != nil {
		if !allowKeyNotFound || err != store.ErrKeyNotFound {
			return nil, err
		}
	}
	pairs := make([]*client.KVPair, 0, len(ps))
	var prefix string
	for _, p := range ps {
		if prefix == "" {
			if strings.HasPrefix(p.Key, "/") {
				if strings.HasPrefix(d.basePath, "/") {
					prefix = d.basePath + "/"
				} else {
					prefix = "/" + d.basePath + "/"
				}
			} else {
				if strings.HasPrefix(d.basePath, "/") {
					prefix = d.basePath[1:] + "/"
				} else {
					prefix = d.basePath + "/"
				}
			}
		}
		k := strings.TrimPrefix(p.Key, prefix)
		pair := &client.KVPair{Key: k, Value: string(p.Value)}
		if d.filter != nil && !d.filter(pair) {
			continue
		}
		pairs = append(pairs, pair)
	}
	d.pairsMu.Lock()
	d.pairs = pairs
	d.pairsMu.Unlock()
	d.RetriesAfterWatchFailed = -1

	go d.watch()
	return d, nil
}

// NewEtcdDiscoveryTemplate returns a new EtcdDiscovery template.
func NewEtcdDiscoveryTemplate(basePath string, etcdAddr []string, allowKeyNotFound bool, options *store.Config) (client.ServiceDiscovery, error) {
	if len(basePath) > 1 && strings.HasSuffix(basePath, "/") {
		basePath = basePath[:len(basePath)-1]
	}

	kv, err := libkv.NewStore(ETCD, etcdAddr, options)
	if err != nil {
		return nil, err
	}

	return NewEtcdDiscoveryStore(basePath, kv, allowKeyNotFound)
}

// Clone clones this ServiceDiscovery with new servicePath.
func (d *EtcdDiscovery) Clone(servicePath string) (client.ServiceDiscovery, error) {
	return NewEtcdDiscoveryStore(d.basePath+"/"+servicePath, d.kv, d.AllowKeyNotFound)
}

// SetFilter sets the filer.
func (d *EtcdDiscovery) SetFilter(filter client.ServiceDiscoveryFilter) {
	d.filter = filter
}

// GetServices returns the servers
func (d *EtcdDiscovery) GetServices() []*client.KVPair {
	d.pairsMu.RLock()
	defer d.pairsMu.RUnlock()
	return d.pairs
}

// WatchService returns a nil chan.
func (d *EtcdDiscovery) WatchService() chan []*client.KVPair {
	d.mu.Lock()
	defer d.mu.Unlock()

	ch := make(chan []*client.KVPair, 10)
	d.chans = append(d.chans, ch)
	return ch
}

func (d *EtcdDiscovery) RemoveWatcher(ch chan []*client.KVPair) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var chans []chan []*client.KVPair
	for _, c := range d.chans {
		if c == ch {
			continue
		}

		chans = append(chans, c)
	}

	d.chans = chans
}

func (d *EtcdDiscovery) watch() {
	defer func() {
		d.kv.Close()
	}()

rewatch:
	for {
		var err error
		var c <-chan []*store.KVPair
		var tempDelay time.Duration

		retry := d.RetriesAfterWatchFailed
		for d.RetriesAfterWatchFailed < 0 || retry >= 0 {
			c, err = d.kv.WatchTree(d.basePath, nil)
			if err != nil {
				if d.RetriesAfterWatchFailed > 0 {
					retry--
				}
				if tempDelay == 0 {
					tempDelay = 1 * time.Second
				} else {
					tempDelay *= 2
				}
				if max := 30 * time.Second; tempDelay > max {
					tempDelay = max
				}
				logger.Warn("can not watchtree (with retry %d, sleep %v): %s: %v", retry, tempDelay, d.basePath, err)
				time.Sleep(tempDelay)
				continue
			}
			break
		}

		if err != nil {
			logger.Error("can't watch %s: %v", d.basePath, err)
			return
		}

		for {
			select {
			case <-d.stopCh:
				return
			case ps, ok := <-c:
				if !ok {
					break rewatch
				}
				var pairs []*client.KVPair // latest servers
				if ps == nil {
					d.pairsMu.Lock()
					d.pairs = pairs
					d.pairsMu.Unlock()
					continue
				}
				var prefix string
				for _, p := range ps {
					if prefix == "" {
						if strings.HasPrefix(p.Key, "/") {
							if strings.HasPrefix(d.basePath, "/") {
								prefix = d.basePath + "/"
							} else {
								prefix = "/" + d.basePath + "/"
							}
						} else {
							if strings.HasPrefix(d.basePath, "/") {
								prefix = d.basePath[1:] + "/"
							} else {
								prefix = d.basePath + "/"
							}
						}
					}
					k := strings.TrimPrefix(p.Key, prefix)
					pair := &client.KVPair{Key: k, Value: string(p.Value)}
					if d.filter != nil && !d.filter(pair) {
						continue
					}
					pairs = append(pairs, pair)
				}
				d.pairsMu.Lock()
				d.pairs = pairs
				d.pairsMu.Unlock()

				d.mu.Lock()
				for _, ch := range d.chans {
					ch := ch
					go func() {
						defer func() {
							recover()
						}()

						select {
						case ch <- pairs:
						case <-time.After(time.Minute):
							logger.Warn("chan is full and new change has been dropped")
						}
					}()
				}
				d.mu.Unlock()
			}
		}
	}
}

func (d *EtcdDiscovery) Close() {
	close(d.stopCh)
}
