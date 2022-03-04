package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hwcer/cosgo/library/logger"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rpcxio/libkv"
	"github.com/rpcxio/libkv/store"
)

const RPCREDIS store.Backend = "RPCREDIS"

var (
	// ErrMultipleEndpointsUnsupported is thrown when there are
	// multiple endpoints specified for Redis
	ErrMultipleEndpointsUnsupported = errors.New("redis: does not support multiple endpoints")

	// ErrTLSUnsupported is thrown when tls config is given
	ErrTLSUnsupported = errors.New("redis does not support tls")

	// ErrAbortTryLock is thrown when a user stops trying to seek the lock
	// by sending a signal to the stop chan, this is used to verify if the
	// operation succeeded
	ErrAbortTryLock = errors.New("redis: lock operation aborted")
)

// Register registers Redis to valkeyrie
func Register() {
	libkv.AddStore(RPCREDIS, New)
}

// New creates a new Redis client given a list
// of endpoints and optional tls config
func New(endpoints []string, options *store.Config) (store.Store, error) {
	var password string
	if len(endpoints) > 1 {
		return nil, ErrMultipleEndpointsUnsupported
	}
	if options != nil && options.TLS != nil {
		return nil, ErrTLSUnsupported
	}
	if options != nil && options.Password != "" {
		password = options.Password
	}

	dbIndex := 0
	if options != nil {
		dbIndex, _ = strconv.Atoi(options.Bucket)
	}

	return newRedis(endpoints, password, dbIndex)
}

func newRedis(endpoints []string, password string, dbIndex int) (*Redis, error) {
	// TODO: use *redis.ClusterClient if we support miltiple endpoints
	client := redis.NewClient(&redis.Options{
		Addr:         endpoints[0],
		DialTimeout:  5 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Password:     password,
		DB:           dbIndex,
	})

	// Listen to Keyspace events
	client.ConfigSet(context.Background(), "notify-keyspace-events", "KEA")

	r := &Redis{
		client: client,
		script: redis.NewScript(luaScript()),
		codec:  defaultCodec{},
		done:   new(sync.WaitGroup),
	}
	r.done.Add(1)
	return r, nil
}

type defaultCodec struct{}

func (c defaultCodec) encode(kv *store.KVPair) (string, error) {
	b, err := json.Marshal(kv)
	return string(b), err
}

func (c defaultCodec) decode(b string, kv *store.KVPair) error {
	return json.Unmarshal([]byte(b), kv)
}

// Redis implements valkeyrie.Store interface with redis backend
type Redis struct {
	client  *redis.Client
	script  *redis.Script
	codec   defaultCodec
	stopped int32
	done    *sync.WaitGroup
}

const (
	noExpiration   = time.Duration(0)
	defaultLockTTL = 60 * time.Second
)

// Put a value at the specified key
func (r *Redis) Put(key string, value []byte, options *store.WriteOptions) error {
	expirationAfter := noExpiration
	if options != nil && options.TTL != 0 {
		expirationAfter = options.TTL
	}

	return r.set(normalize(key), &store.KVPair{
		Key:       key,
		Value:     value,
		LastIndex: sequenceNum(),
	}, expirationAfter)
}

// Get a value given its key
func (r *Redis) Get(key string) (*store.KVPair, error) {
	return r.get(normalize(key))
}
func (r *Redis) set(key string, val *store.KVPair, ttl time.Duration) error {
	valStr, err := r.codec.encode(val)
	if err != nil {
		return err
	}

	return r.client.Set(context.Background(), key, valStr, ttl).Err()
}
func (r *Redis) get(key string) (*store.KVPair, error) {
	reply, err := r.client.Get(context.Background(), key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, store.ErrKeyNotFound
		}
		return nil, err
	}
	val := store.KVPair{}
	if err := r.codec.decode(string(reply), &val); err != nil {
		return nil, err
	}
	return &val, nil
}

// Delete the value at the specified key
func (r *Redis) Delete(key string) error {
	return r.client.Del(context.Background(), normalize(key)).Err()
}

// Exists verify if a Fields exists in the store
func (r *Redis) Exists(key string) (bool, error) {
	i, err := r.client.Exists(context.Background(), normalize(key)).Result()
	if err != nil {
		return false, err
	}
	return i == 1, nil
}

// Watch for changes on a key
// glitch: we use notified-then-retrieve to retrieve *store.KVPair.
// so the responses may sometimes inaccurate
func (r *Redis) Watch(key string, stopCh <-chan struct{}) (<-chan *store.KVPair, error) {
	r.done.Add(1)
	defer r.done.Done()
	watchCh := make(chan *store.KVPair)
	nKey := normalize(key)

	get := getter(func() (interface{}, error) {
		pair, err := r.get(nKey)
		if err != nil {
			return nil, err
		}
		return pair, nil
	})

	push := pusher(func(v interface{}) {
		if val, ok := v.(*store.KVPair); ok {
			watchCh <- val
		}
	})

	sub, err := newSubscribe(r.client, regexWatch(nKey, false))
	if err != nil {
		return nil, err
	}

	go func(sub *subscribe, stopCh <-chan struct{}, get getter, push pusher) {
		r.done.Add(1)
		defer r.done.Done()
		defer sub.Close()
		msgCh := sub.Receive(stopCh)
		if err := watchLoop(msgCh, stopCh, get, push); err != nil {
			logger.Debug("watchLoop in Watch err:%v\n", err)
		}
	}(sub, stopCh, get, push)

	return watchCh, nil
}

func regexWatch(key string, withChildren bool) string {
	var regex string
	if withChildren {
		regex = fmt.Sprintf("__keyspace*:%s*", key)
		// for all database and keys with $key prefix
	} else {
		regex = fmt.Sprintf("__keyspace*:%s", key)
		// for all database and keys with $key
	}
	return regex
}

// getter defines a func type which retrieves data from remote storage
type getter func() (interface{}, error)

// pusher defines a func type which pushes data blob into watch channel
type pusher func(interface{})

func watchLoop(msgCh chan *redis.Message, stopCh <-chan struct{}, get getter, push pusher) error {
	// deliver the original data before we setup any events
	pair, err := get()
	if err != nil {
		return err
	}
	push(pair)

	for m := range msgCh {
		//logger.Debug("watchLoop:%v", m.String())
		// retrieve and send back
		pair, err := get()
		if err != nil && err != store.ErrKeyNotFound {
			return err
		}

		// in case of watching a key that has been expired or deleted return and empty KV
		if err == store.ErrKeyNotFound && (m.Payload == "expire" || m.Payload == "del") {
			push(&store.KVPair{})
		} else {
			push(pair)
		}
	}

	return nil
}

// WatchTree watches for changes on child nodes under
// a given directory
func (r *Redis) WatchTree(directory string, stopCh <-chan struct{}) (<-chan []*store.KVPair, error) {
	r.done.Add(1)
	defer r.done.Done()
	watchCh := make(chan []*store.KVPair)
	nKey := normalize(directory)

	get := getter(func() (interface{}, error) {
		pair, err := r.list(nKey)
		if err != nil {
			return nil, err
		}
		return pair, nil
	})

	push := pusher(func(v interface{}) {
		if _, ok := v.([]*store.KVPair); !ok {
			return
		}
		watchCh <- v.([]*store.KVPair)
	})

	sub, err := newSubscribe(r.client, regexWatch(nKey, true))
	if err != nil {
		return nil, err
	}

	go func(sub *subscribe, stopCh <-chan struct{}, get getter, push pusher) {
		r.done.Add(1)
		defer r.done.Done()
		defer sub.Close()
		msgCh := sub.Receive(stopCh)
		if err := watchLoop(msgCh, stopCh, get, push); err != nil {
			logger.Debug("watchLoop in WatchTree err:%v", err)
		}
	}(sub, stopCh, get, push)

	return watchCh, nil
}

// NewLock creates a lock for a given key.
// The returned Locker is not held and must be acquired
// with `.Lock`. The Value is optional.
func (r *Redis) NewLock(key string, options *store.LockOptions) (store.Locker, error) {
	var (
		value []byte
		ttl   = defaultLockTTL
	)

	if options != nil && options.TTL != 0 {
		ttl = options.TTL
	}
	if options != nil && len(options.Value) != 0 {
		value = options.Value
	}

	return &redisLock{
		redis:    r,
		last:     nil,
		key:      key,
		value:    value,
		ttl:      ttl,
		unlockCh: make(chan struct{}),
	}, nil
}

// List the content of a given prefix
func (r *Redis) List(directory string) ([]*store.KVPair, error) {
	return r.list(normalize(directory))
}

func (r *Redis) list(directory string) ([]*store.KVPair, error) {

	var allKeys []string
	regex := scanRegex(directory) // for all keyed with $directory
	allKeys, err := r.keys(regex)
	if err != nil {
		return nil, err
	}
	// TODO: need to handle when #key is too large
	return r.mget(directory, allKeys...)
}

func (r *Redis) keys(regex string) ([]string, error) {
	const (
		startCursor  = 0
		endCursor    = 0
		defaultCount = 10
	)

	var allKeys []string

	keys, nextCursor, err := r.client.Scan(context.Background(), startCursor, regex, defaultCount).Result()
	if err != nil {
		return nil, err
	}
	allKeys = append(allKeys, keys...)
	for nextCursor != endCursor {
		keys, nextCursor, err = r.client.Scan(context.Background(), nextCursor, regex, defaultCount).Result()
		if err != nil {
			return nil, err
		}

		allKeys = append(allKeys, keys...)
	}
	if len(allKeys) == 0 {
		return nil, store.ErrKeyNotFound
	}
	return allKeys, nil
}

// mget values given their keys
func (r *Redis) mget(directory string, keys ...string) ([]*store.KVPair, error) {
	replies, err := r.client.MGet(context.Background(), keys...).Result()
	if err != nil {
		return nil, err
	}

	pairs := []*store.KVPair{}
	for _, reply := range replies {
		var sreply string
		if _, ok := reply.(string); ok {
			sreply = reply.(string)
		}
		if sreply == "" {
			// empty reply
			continue
		}

		newkv := &store.KVPair{}
		if err := r.codec.decode(sreply, newkv); err != nil {
			return nil, err
		}
		if normalize(newkv.Key) != directory {
			pairs = append(pairs, newkv)
		}
	}
	return pairs, nil
}

// DeleteTree deletes a range of keys under a given directory
// glitch: we list all available keys first and then delete them all
// it costs two operations on redis, so is not atomicity.
func (r *Redis) DeleteTree(directory string) error {
	var allKeys []string
	regex := scanRegex(normalize(directory)) // for all keyed with $directory
	allKeys, err := r.keys(regex)
	if err != nil {
		return err
	}
	return r.client.Del(context.Background(), allKeys...).Err()
}

// AtomicPut is an atomic CAS operation on a single value.
// Pass previous = nil to create a new key.
// we introduced script on this page, so atomicity is guaranteed
func (r *Redis) AtomicPut(key string, value []byte, previous *store.KVPair, options *store.WriteOptions) (bool, *store.KVPair, error) {
	expirationAfter := noExpiration
	if options != nil && options.TTL != 0 {
		expirationAfter = options.TTL
	}

	newKV := &store.KVPair{
		Key:       key,
		Value:     value,
		LastIndex: sequenceNum(),
	}
	nKey := normalize(key)

	// if previous == nil, set directly
	if previous == nil {
		if err := r.setNX(nKey, newKV, expirationAfter); err != nil {
			return false, nil, err
		}
		return true, newKV, nil
	}

	if err := r.cas(
		nKey,
		previous,
		newKV,
		formatSec(expirationAfter),
	); err != nil {
		return false, nil, err
	}
	return true, newKV, nil
}

func (r *Redis) setNX(key string, val *store.KVPair, expirationAfter time.Duration) error {
	valBlob, err := r.codec.encode(val)
	if err != nil {
		return err
	}

	if !r.client.SetNX(context.Background(), key, valBlob, expirationAfter).Val() {
		return store.ErrKeyExists
	}
	return nil
}

func (r *Redis) cas(key string, old, new *store.KVPair, secInStr string) error {
	newVal, err := r.codec.encode(new)
	if err != nil {
		return err
	}

	oldVal, err := r.codec.encode(old)
	if err != nil {
		return err
	}

	return r.runScript(
		cmdCAS,
		key,
		oldVal,
		newVal,
		secInStr,
	)
}

// AtomicDelete is an atomic delete operation on a single value
// the value will be deleted if previous matched the one stored in db
func (r *Redis) AtomicDelete(key string, previous *store.KVPair) (bool, error) {
	if err := r.cad(normalize(key), previous); err != nil {
		return false, err
	}
	return true, nil
}

func (r *Redis) cad(key string, old *store.KVPair) error {
	oldVal, err := r.codec.encode(old)
	if err != nil {
		return err
	}

	return r.runScript(
		cmdCAD,
		key,
		oldVal,
	)
}

// Close the store connection
func (r *Redis) Close() {
	if !atomic.CompareAndSwapInt32(&r.stopped, 0, 1) {
		return
	}
	r.done.Done()
	r.done.Wait()
	r.client.Close()
}

func scanRegex(directory string) string {
	return fmt.Sprintf("%s*", directory)
}

func (r *Redis) runScript(args ...interface{}) error {
	err := r.script.Run(
		context.Background(),
		r.client,
		nil,
		args...,
	).Err()
	if err != nil && strings.Contains(err.Error(), "redis: key is not found") {
		return store.ErrKeyNotFound
	}
	if err != nil && strings.Contains(err.Error(), "redis: value has been changed") {
		return store.ErrKeyModified
	}
	return err
}

func normalize(key string) string {
	return store.Normalize(key)
}

func formatSec(dur time.Duration) string {
	return fmt.Sprintf("%d", int(dur/time.Second))
}

func sequenceNum() uint64 {
	// TODO: use uuid if we concerns collision probability of this number
	return uint64(time.Now().Nanosecond())
}
