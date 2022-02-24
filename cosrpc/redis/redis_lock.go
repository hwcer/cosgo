package redis

import (
	"time"

	"github.com/rpcxio/libkv/store"
)

type redisLock struct {
	redis    *Redis
	last     *store.KVPair
	unlockCh chan struct{}

	key   string
	value []byte
	ttl   time.Duration
}

func (l *redisLock) Lock(stopCh chan struct{}) (<-chan struct{}, error) {
	lockHeld := make(chan struct{})

	success, err := l.tryLock(lockHeld, stopCh)
	if err != nil {
		return nil, err
	}
	if success {
		return lockHeld, nil
	}

	// wait for changes on the key
	watch, err := l.redis.Watch(l.key, stopCh)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case <-stopCh:
			return nil, ErrAbortTryLock
		case <-watch:
			success, err := l.tryLock(lockHeld, stopCh)
			if err != nil {
				return nil, err
			}
			if success {
				return lockHeld, nil
			}
		}
	}
}

// tryLock return true, nil when it acquired and hold the lock
// and return false, nil when it can't lock now,
// and return false, err if any unespected error happened underlying
func (l *redisLock) tryLock(lockHeld, stopChan chan struct{}) (bool, error) {
	success, new, err := l.redis.AtomicPut(
		l.key,
		l.value,
		l.last,
		&store.WriteOptions{
			TTL: l.ttl,
		})
	if success {
		l.last = new
		// keep holding
		go l.holdLock(lockHeld, stopChan)
		return true, nil
	}
	if err != nil && (err == store.ErrKeyNotFound || err == store.ErrKeyModified || err == store.ErrKeyExists) {
		return false, nil
	}
	return false, err
}

func (l *redisLock) holdLock(lockHeld, stopChan chan struct{}) {
	defer close(lockHeld)

	hold := func() error {
		_, new, err := l.redis.AtomicPut(
			l.key,
			l.value,
			l.last,
			&store.WriteOptions{
				TTL: l.ttl,
			})
		if err == nil {
			l.last = new
		}
		return err
	}

	heartbeat := time.NewTicker(l.ttl / 3)
	defer heartbeat.Stop()

	for {
		select {
		case <-heartbeat.C:
			if err := hold(); err != nil {
				return
			}
		case <-l.unlockCh:
			return
		case <-stopChan:
			return
		}
	}
}

func (l *redisLock) Unlock() error {
	l.unlockCh <- struct{}{}

	_, err := l.redis.AtomicDelete(l.key, l.last)
	if err != nil {
		return err
	}
	l.last = nil

	return err
}
