package kv

import (
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/kv"
	_ "github.com/cerana/cerana/pkg/kv/consul" // register consul with pkg/kv
	"github.com/cerana/cerana/provider"
)

// KV is a provider of kv functionality.
type KV struct {
	kv      kv.KV
	config  *Config
	tracker *acomm.Tracker
}

// New creates a new instance of KV.
func New(config *Config, tracker *acomm.Tracker) (*KV, error) {
	addr, err := config.Address()
	if err != nil {
		return nil, err
	}

	k, err := kv.New(addr)
	if err != nil {
		return nil, err
	}

	return &KV{kv: k, config: config, tracker: tracker}, nil
}

// RegisterTasks registers all of KV's task handlers with the server.
func (k *KV) RegisterTasks(server *provider.Server) {
	// simple.go
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
}
