package kv

import (
	"math/rand"
	"sync"

	"github.com/cerana/cerana/pkg/errors"
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
		return nil, errors.Newv("non-existent cookie", map[string]interface{}{"cookie": cookie})
	}

	delete(c.cookies, cookie)
	return ch, nil
}

func (c *chanMap) Peek(cookie uint64) (chan struct{}, error) {
	c.Lock()
	defer c.Unlock()

	ch, exists := c.cookies[cookie]
	if !exists {
		return nil, errors.Newv("non-existent cookie", map[string]interface{}{"cookie": cookie})
	}

	return ch, nil
}

type eKeyMap struct {
	sync.Mutex
	keys map[string]kv.EphemeralKey
}

func newEKeyMap() *eKeyMap {
	return &eKeyMap{
		keys: map[string]kv.EphemeralKey{},
	}
}

func (e *eKeyMap) Destroy(key string) error {
	e.Lock()
	defer e.Unlock()

	eKey, ok := e.keys[key]
	if !ok {
		return errors.Newv("unknown ephemeral key", map[string]interface{}{"key": key})
	}

	if err := eKey.Destroy(); err != nil {
		return err
	}

	delete(e.keys, key)

	return nil
}

func (e *eKeyMap) Get(key string) kv.EphemeralKey {
	e.Lock()
	defer e.Unlock()

	return e.keys[key]
}

// Add stores the EphemeralKey in the map, overwriting existent values
func (e *eKeyMap) Add(key string, eKey kv.EphemeralKey) {
	e.Lock()
	defer e.Unlock()

	e.keys[key] = eKey
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
		return 0, errors.Newv("failed to create random cookie, try again", map[string]interface{}{"lock": lock})
	}
	return uint64(cookie), nil
}

func (l *lockMap) Get(cookie uint64) (kv.Lock, error) {
	l.Lock()
	defer l.Unlock()

	lock, exists := l.cookies[cookie]
	if !exists {
		return nil, errors.Newv("non-existent cookie", map[string]interface{}{"cookie": cookie})
	}

	delete(l.cookies, cookie)
	return lock, nil
}

func (l *lockMap) Peek(cookie uint64) (kv.Lock, error) {
	l.Lock()
	defer l.Unlock()

	lock, exists := l.cookies[cookie]
	if !exists {
		return nil, errors.Newv("non-existent cookie", map[string]interface{}{"cookie": cookie})
	}

	return lock, nil
}
