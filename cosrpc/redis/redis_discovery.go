package redis

import (
	"github.com/hwcer/cosgo/logger"
	"github.com/smallnest/rpcx/client"
	"strings"
	"sync"
	"time"

	"github.com/rpcxio/libkv"
	"github.com/rpcxio/libkv/store"
)

func init() {
	Register()
}

// RedisDiscovery is a redis service discovery.
// It always returns the registered servers in redis.
type RedisDiscovery struct {
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
	stopCh           chan struct{}
}

// NewRedisDiscovery returns a new RedisDiscovery.
func NewRedisDiscovery(basePath string, servicePath string, redisAddr []string, allowKeyNotFound bool, options *store.Config) (client.ServiceDiscovery, error) {
	kv, err := libkv.NewStore(RPCREDIS, redisAddr, options)
	if err != nil {
		return nil, err
	}

	return NewRedisDiscoveryStore(basePath+"/"+servicePath, kv, allowKeyNotFound)
}

// NewRedisDiscoveryStore return a new RedisDiscovery with specified store.
func NewRedisDiscoveryStore(basePath string, kv store.Store, allowKeyNotFound bool) (client.ServiceDiscovery, error) {
	if len(basePath) > 1 && strings.HasSuffix(basePath, "/") {
		basePath = strings.TrimSuffix(basePath, "/")
	}

	d := &RedisDiscovery{basePath: basePath, kv: kv}
	d.stopCh = make(chan struct{})

	ps, err := kv.List(basePath)
	if err != nil {
		if !allowKeyNotFound || err != store.ErrKeyNotFound {
			return nil, err
		}
	}
	d.setPairs(ps)
	d.RetriesAfterWatchFailed = -1
	d.AllowKeyNotFound = allowKeyNotFound
	go d.watch()
	return d, nil
}

// NewRedisDiscoveryTemplate returns a new RedisDiscovery template.
func NewRedisDiscoveryTemplate(basePath string, etcdAddr []string, allowKeyNotFound bool, options *store.Config) (client.ServiceDiscovery, error) {
	if len(basePath) > 1 && strings.HasSuffix(basePath, "/") {
		basePath = basePath[:len(basePath)-1]
	}
	kv, err := libkv.NewStore(RPCREDIS, etcdAddr, options)
	if err != nil {
		return nil, err
	}

	return NewRedisDiscoveryStore(basePath, kv, allowKeyNotFound)
}

// Clone clones this ServiceDiscovery with new servicePath.
func (d *RedisDiscovery) Clone(servicePath string) (client.ServiceDiscovery, error) {
	return NewRedisDiscoveryStore(d.basePath+"/"+servicePath, d.kv, d.AllowKeyNotFound)
}

// SetFilter sets the filer.
func (d *RedisDiscovery) SetFilter(filter client.ServiceDiscoveryFilter) {
	d.filter = filter
}

// GetServices returns the servers
func (d *RedisDiscovery) GetServices() []*client.KVPair {
	d.pairsMu.RLock()
	defer d.pairsMu.RUnlock()

	return d.pairs
}

// WatchService returns a nil chan.
func (d *RedisDiscovery) WatchService() chan []*client.KVPair {
	d.mu.Lock()
	defer d.mu.Unlock()

	ch := make(chan []*client.KVPair, 10)
	d.chans = append(d.chans, ch)
	return ch
}

func (d *RedisDiscovery) RemoveWatcher(ch chan []*client.KVPair) {
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

func (d *RedisDiscovery) watch() {
	defer func() {
		d.kv.Close()
	}()

	for {
		var err error
		var c <-chan []*store.KVPair
		var tempDelay time.Duration

		retry := d.RetriesAfterWatchFailed
		for d.RetriesAfterWatchFailed < 0 || retry >= 0 {
			c, err = d.kv.WatchTree(d.basePath, d.stopCh)
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

	readChanges:
		for {
			select {
			case <-d.stopCh:
				return
			case ps, ok := <-c:
				if !ok {
					break readChanges
				}
				var pairs []*client.KVPair // latest servers
				if ps == nil {
					d.pairsMu.Lock()
					d.pairs = pairs
					d.pairsMu.Unlock()
					continue
				} else {
					pairs = d.setPairs(ps)
				}
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

		logger.Warn("chan is closed and will rewatch")
	}
}

func (d *RedisDiscovery) Close() {
	close(d.stopCh)
}

func (d *RedisDiscovery) prefix(key string) (prefix string) {
	if strings.HasPrefix(key, "/") {
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
	return
}

func (d *RedisDiscovery) setPairs(ps []*store.KVPair) []*client.KVPair {
	var prefix string
	var pairs []*client.KVPair // latest servers
	for _, p := range ps {
		if prefix == "" {
			prefix = d.prefix(p.Key)
		}
		if p.Key == prefix[:len(prefix)-1] {
			continue
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
	return pairs
}
