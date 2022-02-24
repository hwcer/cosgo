package updater

import (
	"sync"
)

func NewPool() Pool {
	i := Pool{pool: sync.Pool{}}
	i.pool.New = func() (u interface{}) {
		u, _, _ = New("")
		return
	}
	return i
}

type Pool struct {
	pool sync.Pool
}

func (s *Pool) Acquire(uid string) (u *Updater) {
	u = s.pool.Get().(*Updater)
	u.reset(uid)
	return u
}

func (s *Pool) Release(u *Updater) {
	u.release()
	s.pool.Put(u)
}
