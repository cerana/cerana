package kv

import (
	"errors"
	"math/rand"
	"sync"

	"github.com/cerana/cerana/pkg/kv"
)

type chanMap struct {
	sync.Mutex
	cookies map[uint64]chan struct{}
}

func newChanMap() *chanMap {
	return &chanMap{cookies: map[uint64]chan struct{}{}}
}

func (c *chanMap) Add(ch chan struct{}) (uint64, error) {
	cookie := rand.Int63()
	exists := false

	c.Lock()
	for i := 0; i < 5; i++ {
		if _, exists = c.cookies[uint64(cookie)]; exists {
			cookie = rand.Int63()
			continue
		}
		c.cookies[uint64(cookie)] = ch
		break
	}
	c.Unlock()

	if exists {
		return 0, errors.New("failed to create random cookie, try again")
	}
	return uint64(cookie), nil
}

func (c *chanMap) Get(cookie uint64) (chan struct{}, error) {
	c.Lock()
	defer c.Unlock()

	ch, exists := c.cookies[cookie]
	if !exists {
		return nil, errors.New("non-existent cookie")
	}

	delete(c.cookies, cookie)
	return ch, nil
}

func (c *chanMap) Peek(cookie uint64) (chan struct{}, error) {
	c.Lock()
	defer c.Unlock()

	ch, exists := c.cookies[cookie]
	if !exists {
		return nil, errors.New("non-existent cookie")
	}

	return ch, nil
}

type lockMap struct {
	sync.Mutex
	cookies map[uint64]kv.Lock
}

func newLockMap() *lockMap {
	return &lockMap{cookies: map[uint64]kv.Lock{}}
}

func (l *lockMap) Add(lock kv.Lock) (uint64, error) {
	cookie := rand.Int63()
	exists := false

	l.Lock()
	for i := 0; i < 5; i++ {
		if _, exists = l.cookies[uint64(cookie)]; exists {
			cookie = rand.Int63()
			continue
		}
		l.cookies[uint64(cookie)] = lock
		break
	}
	l.Unlock()

	if exists {
		return 0, errors.New("failed to create random cookie, try again")
	}
	return uint64(cookie), nil
}

func (l *lockMap) Get(cookie uint64) (kv.Lock, error) {
	l.Lock()
	defer l.Unlock()

	lock, exists := l.cookies[cookie]
	if !exists {
		return nil, errors.New("non-existent cookie")
	}

	delete(l.cookies, cookie)
	return lock, nil
}

func (l *lockMap) Peek(cookie uint64) (kv.Lock, error) {
	l.Lock()
	defer l.Unlock()

	lock, exists := l.cookies[cookie]
	if !exists {
		return nil, errors.New("non-existent cookie")
	}

	return lock, nil
}
