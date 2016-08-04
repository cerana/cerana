// Package kv abstracts a distributed/clustered kv store for use with cerana
// kv does not aim to be a full-featured generic kv abstraction, but can be useful anyway.
// Only implementors imported by users will be available at runtime.
// See documentation of KV for handled operations.
package kv

import (
	"fmt"
	"net/url"
	"sync"
	"time"
)

// Value represents the value stored in a key, including the last modification index of the key
type Value struct {
	Data  []byte `json:"data"`
	Index uint64 `json:"index"`
}

// EventType is used to describe actions on watch events
type EventType int

//go:generate sh -c "stringer -type=EventType && sed -i 's#_EventType_name#eventTypeName#g;s#_EventType_index#eventTypeIndex#g' eventtype_string.go"

const (
	// None indicates no event, should induce a panic if ever seen
	None EventType = iota
	// Create indicates a new key being added
	Create
	// Delete indicates a key being deleted
	Delete
	// Update indicates a key being modified, the contents of the key are not taken into account
	Update
)

var types = map[EventType]string{
	None:   "None",
	Create: "Create",
	Delete: "Delete",
	Update: "Update",
}

// Event represents an action occurring to a watched key or prefix
type Event struct {
	Key  string    `json:"key"`
	Type EventType `json:"type"`
	Value
}

var register = struct {
	sync.RWMutex
	kvs map[string]func(string) (KV, error)
}{
	kvs: map[string]func(string) (KV, error){},
}

// Register is called by KV implementors to register their scheme to be used with New
func Register(name string, fn func(string) (KV, error)) {
	register.Lock()
	defer register.Unlock()

	if _, dup := register.kvs[name]; dup {
		panic("kv: Register called twice for " + name)
	}
	register.kvs[name] = fn
}

// New will return a KV implementation according to the connection string addr.
// The parameter addr may be the empty string or a valid URL.
// The special `http` and `https` schemes are deemed generic, the first implementation that supports it will be used.
// Otherwise the scheme portion of the URL will be used to select the exact implementation to instantiate.
func New(addr string) (KV, error) {
	var u *url.URL
	if addr == "" {
		u = &url.URL{Scheme: "http"}
	} else {
		var err error
		u, err = url.Parse(addr)
		if err != nil {
			return nil, err
		}
	}

	register.RLock()
	defer register.RUnlock()

	constructor := register.kvs[u.Scheme]
	if constructor != nil {
		kv, err := constructor(addr)
		if err != nil {
			return nil, err
		}
		if err = kv.Ping(); err != nil {
			return nil, err
		}
		return kv, nil
	} else if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unknown kv store %s (forgotten import?)", u.Scheme)
	}

	for _, constructor := range register.kvs {
		kv, err := constructor(addr)
		if err != nil {
			return nil, err
		}
		// scheme was http(s) so an error just means we tried to connect
		// to an incompatible cluster
		if err := kv.Ping(); err == nil {
			return kv, nil
		}
	}
	return nil, fmt.Errorf("unknown kv store")
}

// Lock represents a locked key in the distributed key value store.
// The value stored in key is managed by lock and may contain private implementation data and should not be fetched out-of-band
type Lock interface {
	// Renew renews the lock, it should be called before attempting any operation on whatever is being protected
	Renew() error
	// Unlock unlocks and invalidates the lock
	Unlock() error
}

// EphemeralKey represents a key that will disappear once the timeout used to instantiate it has lapsed.
type EphemeralKey interface {
	// Set will first renew the ttl then set the value of key, it is an error if the ttl has expired since last renewal
	Set(value string) error
	// Renew renews the key ttl
	Renew() error
	// Destroy will delete the key without having to wait for expiration via TTL
	Destroy() error
}

// KV is the interface for distributed key value store interaction
type KV interface {
	Delete(string, bool) error
	Get(string) (Value, error)
	GetAll(string) (map[string]Value, error)
	Keys(string) ([]string, error)
	Set(string, string) error

	// Atomic operations
	// Update will set key=value while ensuring that newer values are not clobbered
	Update(string, Value) (uint64, error)
	// Remove will delete key only if it has not been modified since index
	Remove(string, uint64) error

	// IsKeyNotFound is a helper to determine if the error is a key not found error
	IsKeyNotFound(error) bool

	// Watch returns channels for watching prefixes for _future_ events.
	// stop *must* always be closed by callers
	// Note: replaying events in history is not guaranteed to be possible.
	Watch(string, uint64, chan struct{}) (chan Event, chan error, error)

	// EphemeralKey creates a key that will be deleted if the ttl expires
	EphemeralKey(string, time.Duration) (EphemeralKey, error)

	// Lock creates a new lock, it blocks until the lock is acquired.
	Lock(string, time.Duration) (Lock, error)

	// Ping verifies communication with the cluster
	Ping() error
}
