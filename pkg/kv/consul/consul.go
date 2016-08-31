package consul

import (
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/pkg/kv"
	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/watch"
)

var err404 = errors.Cause(errors.New("key not found"))

func init() {
	kv.Register("consul", New)
}

type ckv struct {
	c      *consul.KV
	client *consul.Client
	config *consul.Config
}

// New instantiates a consul kv implementation.
// The parameter addr may be the empty string or a valid URL.
// If addr is not empty it must be a valid URL with schemes http, https or consul; consul is synonymous with http.
// If addr is the empty string the consul client will connect to the default address, which may be influenced by the environment.
func New(addr string) (kv.KV, error) {
	config := consul.DefaultConfig()
	if addr == "" {
		addr = config.Scheme + "://" + config.Address
	} else {
		u, err := url.Parse(addr)
		if err != nil {
			return nil, errors.Wrapv(err, map[string]interface{}{"addr": addr}, "failed to parse addr")
		}

		if u.Scheme != "consul" {
			config.Scheme = u.Scheme
		}
		config.Address = u.Host
	}

	client, err := consul.NewClient(config)
	if err != nil {
		return nil, errors.Wrapv(err, map[string]interface{}{"config": config}, "failed to create client")
	}

	return &ckv{c: client.KV(), client: client, config: config}, nil
}

func (c *ckv) Delete(key string, recurse bool) error {
	var err error
	if recurse {
		if key != "" && !strings.HasSuffix(key, "/") {
			key += "/"
		}
		_, err = c.c.DeleteTree(key, nil)
		err = errors.Wrap(err)
	} else {
		_, err = c.c.Delete(key, nil)
		err = errors.Wrap(err)
	}
	return errors.Wrapv(err, map[string]interface{}{"key": key, "recurse": recurse})
}

func (c *ckv) Get(key string) (kv.Value, error) {
	kvp, _, err := c.c.Get(key, nil)
	if err != nil {
		return kv.Value{}, errors.Wrapv(err, map[string]interface{}{"key": key})
	}
	if kvp == nil || kvp.Value == nil {
		return kv.Value{}, errors.Wrapv(err404, map[string]interface{}{"key": key})
	}
	return kv.Value{Data: kvp.Value, Index: kvp.ModifyIndex}, nil
}

func (c *ckv) GetAll(prefix string) (map[string]kv.Value, error) {
	pairs, _, err := c.c.List(prefix, nil)
	if err != nil {
		return nil, errors.Wrapv(err, map[string]interface{}{"prefix": prefix})
	}
	many := make(map[string]kv.Value, len(pairs))
	for _, kvp := range pairs {
		many[kvp.Key] = kv.Value{Data: kvp.Value, Index: kvp.ModifyIndex}
	}
	return many, nil
}

func (c *ckv) Keys(key string) ([]string, error) {
	if !strings.HasSuffix(key, "/") {
		key += "/"
	}
	keys, _, err := c.c.Keys(key, "/", nil)
	return keys, errors.Wrapv(err, map[string]interface{}{"key": key})
}

func (c *ckv) Set(key, value string) error {
	_, err := c.c.Put(&consul.KVPair{Key: key, Value: []byte(value)}, nil)
	return errors.Wrapv(err, map[string]interface{}{"key": key, "value": value})
}

func (c *ckv) cas(key string, value kv.Value) error {
	kvp := consul.KVPair{
		Key:         key,
		Value:       value.Data,
		ModifyIndex: value.Index,
	}

	valid, _, err := c.c.CAS(&kvp, nil)
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"key": key, "value": value})
	}

	if !valid {
		// TODO(mm) better error
		return errors.Newv("CAS failed", map[string]interface{}{"key": key, "value": value})
	}

	return nil
}

// Update is racy with other modifiers since the consul KV API does not return the new modified index.
// See https://github.com/hashicorp/consul/issues/304
func (c *ckv) Update(key string, value kv.Value) (uint64, error) {
	// TODO(mmlb): setup a key watch before this update to get the new modified index
	// can be if cas works then the watched index returned is valid
	err := c.cas(key, value)
	if err != nil {
		return 0, err
	}

	v, err := c.Get(key)
	return v.Index, err
}

func (c *ckv) Remove(key string, index uint64) error {
	ok, _, err := c.c.DeleteCAS(&consul.KVPair{Key: key, ModifyIndex: index}, nil)
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"key": key, "index": index})
	}

	if !ok {
		err = errors.Newv("failed to delete atomically", map[string]interface{}{"key": key, "index": index})
	}

	return err
}

func (c *ckv) IsKeyNotFound(err error) bool {
	return errors.Cause(err) == err404
}

func (c *ckv) Watch(prefix string, lastIndex uint64, stop chan struct{}) (chan kv.Event, chan error, error) {
	wp, err := watch.Parse(map[string]interface{}{
		"type":   "keyprefix",
		"prefix": prefix,
	})
	if err != nil {
		return nil, nil, errors.Wrapv(err, map[string]interface{}{"type": "keyprefix", "prefix": prefix})
	}

	events := make(chan kv.Event)
	errs := make(chan error)

	lastState := map[string]uint64{}
	wp.Handler = func(newIndex uint64, data interface{}) {
		newState := map[string]uint64{}
		for _, kvp := range data.(consul.KVPairs) {
			newState[kvp.Key] = kvp.ModifyIndex

			// from before time we care about, so not an Event
			if kvp.ModifyIndex <= lastIndex {
				delete(lastState, kvp.Key)
				continue
			}

			event := kv.Event{
				Key:  kvp.Key,
				Type: kv.Update,
				Value: kv.Value{
					Data:  kvp.Value,
					Index: kvp.ModifyIndex,
				},
			}

			if _, ok := lastState[kvp.Key]; !ok {
				event.Type = kv.Create
			} else {
				delete(lastState, kvp.Key)
			}
			events <- event
		}

		// anything left over in lastState has not been found in
		// newState so it must have been deleted
		for key, index := range lastState {
			events <- kv.Event{
				Key:  key,
				Type: kv.Delete,
				Value: kv.Value{
					Index: index,
				},
			}
		}

		lastState = newState
		lastIndex = newIndex
	}

	go func() {
		<-stop
		wp.Stop()
		close(events)
		close(errs)
	}()
	go func() {
		err = wp.Run(c.config.Address)
		if err != nil {
			errs <- errors.Wrapv(err, map[string]interface{}{"addr": c.config.Address})
		}
	}()

	return events, errs, nil
}

type lock struct {
	sessions *consul.Session
	kv       *consul.KV
	session  string
	key      string
}

func (c *ckv) lock(key string, ttl time.Duration, behavior string) (string, error) {
	sEntry := &consul.SessionEntry{
		TTL:      ttl.String(),
		Behavior: behavior,
	}

	session, _, err := c.client.Session().Create(sEntry, nil)
	if err != nil {
		return "", errors.Wrapv(err, map[string]interface{}{"entry": sEntry}, "failed to create session")
	}

	ok, _, err := c.c.Acquire(&consul.KVPair{Key: key, Session: session}, nil)
	if err != nil {
		return "", errors.Wrapv(err, map[string]interface{}{"key": key, "session": session})
	}
	if !ok {
		return "", errors.Newv("lock held by another client", map[string]interface{}{"key": key})
	}
	return session, nil
}

func (c *ckv) Lock(key string, ttl time.Duration) (kv.Lock, error) {
	session, err := c.lock(key, ttl, consul.SessionBehaviorRelease)
	if err != nil {
		return nil, err
	}

	return &lock{sessions: c.client.Session(), session: session, kv: c.c, key: key}, nil
}

func (c *ckv) EphemeralKey(key string, ttl time.Duration) (kv.EphemeralKey, error) {
	session, err := c.lock(key, ttl, consul.SessionBehaviorDelete)
	if err != nil {
		return nil, err
	}

	return &ekey{kv: c, lock: lock{kv: c.c, sessions: c.client.Session(), session: session, key: key}}, nil
}

func (l *lock) Renew() error {
	entry, _, err := l.sessions.Renew(l.session, nil)
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"key": l.key, "session": l.session})
	}
	if entry == nil {
		return errors.Newv("lock not held", map[string]interface{}{"key": l.key, "session": l.session})
	}
	return nil
}

func (l *lock) unlock() error {
	err := l.Renew()
	if err != nil {
		return err
	}
	ok, _, err := l.kv.Release(&consul.KVPair{Key: l.key, Session: l.session}, nil)
	if err != nil {
		return err
	}
	if !ok {
		return errors.Newv("lock not held", map[string]interface{}{"key": l.key, "session": l.session})
	}
	return nil
}

func (l *lock) Unlock() error {
	return l.unlock()
}

// Ping verifies communication with the cluster
func (c *ckv) Ping() error {
	_, err := c.client.Agent().NodeName()
	return errors.Wrap(err)
}

func (c *ckv) IsLeader() (bool, error) {
	leaderAddr, err := c.client.Status().Leader()
	if err != nil {
		return false, errors.Wrap(err)
	}

	leader, _, err := net.SplitHostPort(leaderAddr)
	if err != nil {
		return false, errors.Wrap(err)
	}

	self, err := c.client.Agent().Self()
	if err != nil {
		return false, errors.Wrap(err)
	}

	ctx := map[string]interface{}{"self": self}
	member, ok := self["Member"]
	if !ok {
		return false, errors.Newv(`missing "Member" key in description of self`, ctx)
	}

	addrI, ok := member["Addr"]
	if !ok {
		return false, errors.Newv(`missing "Addr" key from self.Member`, ctx)
	}
	addr, ok := addrI.(string)
	if !ok {
		return false, errors.Newv(`"self.Member.Addr" value is an unexpected type`, ctx)
	}

	return addr == leader, nil
}

type ekey struct {
	kv *ckv
	lock
}

func (e *ekey) Set(value string) error {
	err := e.Renew()
	if err != nil {
		return err
	}
	return e.kv.Set(e.key, value)
}

func (e *ekey) Destroy() error {
	return e.unlock()
}
