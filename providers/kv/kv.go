package kv

import (
	"sync"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/kv"
	_ "github.com/cerana/cerana/pkg/kv/consul" // register consul with pkg/kv
	"github.com/cerana/cerana/provider"
)

// KV is a provider of kv functionality.
type KV struct {
	config  *Config
	tracker *acomm.Tracker

	mu sync.RWMutex
	kv kv.KV
}

// Value represents the value stored in a key, including the last modification index of the key
type Value kv.Value

type eKVDown string

func (e eKVDown) Temporary() bool {
	return true
}

func (e eKVDown) Error() string {
	return string(e)
}

// errorKVDown indicates that KV has not connected to the KV store yet
const errorKVDown = eKVDown("kv store is down")

// New creates a new instance of KV.
func New(config *Config, tracker *acomm.Tracker) (*KV, error) {
	addr, err := config.Address()
	if err != nil {
		return nil, err
	}

	KV := &KV{config: config, tracker: tracker}
	go func() {
		for {
			k, err := kv.New(addr)
			if err == nil {
				KV.mu.Lock()
				KV.kv = k
				KV.mu.Unlock()
				return

			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
	return KV, nil
}

func (k *KV) kvDown() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.kv == nil
}

// RegisterTasks registers all of KV's task handlers with the server.
func (k *KV) RegisterTasks(server *provider.Server) {
	// basic.go
	server.RegisterTask("kv-is-leader", k.isLeader)
	server.RegisterTask("kv-delete", k.delete)
	server.RegisterTask("kv-get", k.get)
	server.RegisterTask("kv-getAll", k.getAll)
	server.RegisterTask("kv-keys", k.keys)
	server.RegisterTask("kv-set", k.set)

	// cas.go
	server.RegisterTask("kv-remove", k.remove)
	server.RegisterTask("kv-update", k.update)

	// watch.go
	server.RegisterTask("kv-watch", k.watch)
	server.RegisterTask("kv-stop", k.stop)

	// lock.go
	server.RegisterTask("kv-lock", k.lock)
	server.RegisterTask("kv-renew", k.renew)
	server.RegisterTask("kv-unlock", k.unlock)

	// ekey.go
	server.RegisterTask("kv-ephemeral-set", k.eset)
	server.RegisterTask("kv-ephemeral-destroy", k.edestroy)
}
