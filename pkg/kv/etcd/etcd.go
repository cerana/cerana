package etcd

import (
	"errors"
	"net/url"
	"time"

	"github.com/cerana/cerana/pkg/kv"
	etcdErr "github.com/coreos/etcd/error"
	"github.com/coreos/go-etcd/etcd"
)

func init() {
	kv.Register("etcd", New)
}

type ekv struct {
	e *etcd.Client
}

// New instantiates an etcd kv implementation.
// The parameter addr may be the empty string or a valid URL.
// If addr is not empty it must be a valid URL with schemes http, https or etcd; etcd is synonymous with http.
// If addr is the empty string the etcd client will connect to the default address.
func New(addr string) (kv.KV, error) {
	// allow addrs as passed into NewClient to be len == 0, so that etcd
	// will connect to the default address
	addrs := make([]string, 0, 1)
	if addr != "" {
		u, err := url.Parse(addr)
		if err != nil {
			return nil, err
		}

		if u.Scheme == "etcd" {
			u.Scheme = "http"
		}
		addr = u.Scheme + "://" + u.Host
		addrs = append(addrs, addr)
	}
	return &ekv{e: etcd.NewClient(addrs)}, nil
}

func (e *ekv) Delete(key string, recurse bool) error {
	_, err := e.e.Delete(key, recurse)
	if err != nil && e.IsKeyNotFound(err) {
		err = nil
	}
	return err
}

func (e *ekv) Get(key string) (kv.Value, error) {
	resp, err := e.e.Get(key, false, false)
	if err != nil {
		return kv.Value{}, err
	}

	if resp.Node.Dir {
		return kv.Value{}, errors.New("key is a directory")
	}

	return kv.Value{Data: []byte(resp.Node.Value), Index: resp.Node.ModifiedIndex}, nil
}

func (e *ekv) GetAll(prefix string) (map[string]kv.Value, error) {
	resp, err := e.e.Get(prefix, false, true)
	if err != nil {
		return nil, err
	}

	if !resp.Node.Dir {
		return map[string]kv.Value{
			resp.Node.Key: {Data: []byte(resp.Node.Value), Index: resp.Node.ModifiedIndex},
		}, nil
	}

	many := map[string]kv.Value{}
	var recursive func(etcd.Nodes)
	recursive = func(nodes etcd.Nodes) {
		for _, node := range nodes {
			if node.Dir {
				recursive(node.Nodes)
			} else {
				many[node.Key] = kv.Value{Data: []byte(node.Value), Index: node.ModifiedIndex}
			}
		}
	}
	recursive(resp.Node.Nodes)

	return many, nil
}

func (e *ekv) Keys(key string) ([]string, error) {
	resp, err := e.e.Get(key, true, false)
	if err != nil {
		return nil, err
	}

	if !resp.Node.Dir {
		return nil, errors.New("key is not a directory")
	}

	nodes := resp.Node.Nodes
	keys := make([]string, len(nodes))
	for i := range nodes {
		keys[i] = nodes[i].Key
	}

	return keys, err
}

func (e *ekv) Set(key, value string) error {
	_, err := e.e.Set(key, value, 0)
	return err
}

func (e *ekv) Update(key string, value kv.Value) (uint64, error) {
	var err error
	var resp *etcd.Response
	if value.Index == 0 {
		resp, err = e.e.Create(key, string(value.Data), 0)
	} else {
		resp, err = e.e.CompareAndSwap(key, string(value.Data), 0, "", value.Index)
	}
	if err != nil {
		return 0, err
	}
	return resp.Node.ModifiedIndex, nil
}

func (e *ekv) Remove(key string, index uint64) error {
	_, err := e.e.CompareAndDelete(key, "", index)
	return err
}

func (e *ekv) IsKeyNotFound(err error) bool {
	eErr, ok := err.(*etcd.EtcdError)
	return ok && eErr.ErrorCode == etcdErr.EcodeKeyNotFound
}

func (e *ekv) isKeyExists(err error) bool {
	eErr, ok := err.(*etcd.EtcdError)
	return ok && eErr.ErrorCode == etcdErr.EcodeNodeExist
}

var typeE2KV = map[string]kv.EventType{
	"compareAndSwap": kv.Update,
	"create":         kv.Create,
	"delete":         kv.Delete,
	"set":            kv.Update,
}

func (e *ekv) Watch(prefix string, index uint64, stop chan struct{}) (chan kv.Event, chan error, error) {
	bStop := make(chan bool)
	go func() {
		<-stop
		bStop <- true
	}()

	responses := make(chan *etcd.Response)
	events := make(chan kv.Event)
	go func() {
		for resp := range responses {
			events <- kv.Event{
				Type: typeE2KV[resp.Action],
				Key:  resp.Node.Key,
				Value: kv.Value{
					Data:  []byte(resp.Node.Value),
					Index: resp.Node.ModifiedIndex,
				},
			}
		}
	}()

	errors := make(chan error)
	go func() {
		_, err := e.e.Watch(prefix, index, true, responses, bStop)
		if err != nil && err != etcd.ErrWatchStoppedByUser {
			errors <- err
		}
	}()

	return events, errors, nil
}

type lock struct {
	client *etcd.Client
	key    string
	ttl    time.Duration
	index  uint64
}

func (e *ekv) Lock(key string, ttl time.Duration) (kv.Lock, error) {
	if key == "" {
		return nil, errors.New("missing key")
	}

	lock := &lock{client: e.e, key: key, ttl: ttl}

	// Since etcd doesn't really support a lock we need a way discover if a key/lock is held.
	// The safest way to do that is to save something in the kv store with the data, atomically.
	// *cough* something something transactions something *cough*
	//
	// so we prefix the data with locked=true: or locked=false:
	// the locked=true case is to guard against some one trying to take the lock when it is currently locked
	// this is guarded by trying to do a CAS where the previous data was locked=false
	//
	// alternatively we can append locked=true/false to the end this way
	// lock users can json unmarshal the value without having to `Get`. I
	// kind of like the accessors for locks though.

	resp, err := e.e.Create(key, "locked=true", uint64(ttl.Seconds()))
	if err == nil {
		lock.index = resp.Node.ModifiedIndex
		return lock, nil
	} else if !e.isKeyExists(err) {
		return nil, err
	}

	// don't clobber the actual value
	v, err := e.Get(key)
	if err != nil {
		return nil, err
	}

	value := string(v.Data)
	if value != "locked=true" || value != "locked=false" {
		return nil, errors.New("key does not contain a valid Lock value")
	}

	resp, err = e.e.CompareAndSwap(key, "locked=true", uint64(ttl.Seconds()), "locked=false", v.Index)
	if err != nil {
		return nil, err
	}

	lock.index = resp.Node.ModifiedIndex
	return lock, nil
}

func (l *lock) Renew() error {
	resp, err := l.client.CompareAndSwap(l.key, "locked=true", uint64(l.ttl.Seconds()), "", l.index)
	if err != nil {
		return err
	}

	l.index = resp.Node.ModifiedIndex
	return nil
}

func (l *lock) Unlock() error {
	err := l.Renew()
	if err != nil {
		// trying to unlock a lock we don't hold is a logic error
		return err
	}

	_, err = l.client.CompareAndSwap(l.key, "locked=false", uint64(l.ttl.Seconds()), "", l.index)
	if err != nil {
		return err
	}

	l.index = 0
	return nil
}

func (e *ekv) Ping() error {
	if !e.e.SyncCluster() {
		return errors.New("can not reach cluster")
	}
	return nil
}

type eKey struct {
	client *etcd.Client
	key    string
	value  string
	ttl    uint64
}

func (e *ekv) EphemeralKey(key string, ttl time.Duration) (kv.EphemeralKey, error) {
	return eKey{client: e.e, key: key, ttl: uint64(ttl.Seconds())}, nil
}

func (e eKey) Set(value string) error {
	_, err := e.client.Set(e.key, value, e.ttl)
	return err
}

func (e eKey) Renew() error {
	if e.value == "" {
		resp, err := e.client.Get(e.key, false, false)
		if err != nil {
			return err
		}
		e.value = resp.Node.Value
	}
	_, err := e.client.Set(e.key, e.value, e.ttl)
	return err
}

func (e eKey) Destroy() error {
	_, err := e.client.Delete(e.key, false)
	return err
}
