package clusterconf_test

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/provider"
	"github.com/mistifyio/lochness/pkg/kv"
)

type KVP struct {
	config *provider.Config
	Data   map[string]kv.Value
}

func NewKVP(config *provider.Config) *KVP {
	return &KVP{
		config: config,
		Data:   make(map[string]kv.Value),
	}
}

func (k *KVP) RegisterTasks(server provider.Server) {
	server.RegisterTask("kv-getAll", k.GetAll)
	server.RegisterTask("kv-get", k.Get)
	server.RegisterTask("kv-delete", k.Delete)
	server.RegisterTask("kv-update", k.Update)
	server.RegisterTask("kv-ephemeral", k.Ephemeral)
}

type GetArgs struct {
	Key string `json:"key"`
}

func (k *KVP) GetAll(req *acomm.Request) (interface{}, *url.URL, error) {
	var args GetArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	values := make(map[string]kv.Value)
	for key, value := range k.Data {
		if strings.HasPrefix(key, args.Key) {
			values[key] = value
		}
	}

	return values, nil, nil
}

func (k *KVP) Get(req *acomm.Request) (interface{}, *url.URL, error) {
	var args GetArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	value, ok := k.Data[args.Key]
	if !ok {
		return nil, nil, errors.New("key not found")
	}
	return value, nil, nil
}

type DeleteArgs struct {
	Key     string `json:"key"`
	Recurse bool   `json:"recurse"`
}

func (k *KVP) Delete(req *acomm.Request) (interface{}, *url.URL, error) {
	var args DeleteArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	found := false
	for key := range k.Data {
		if strings.HasPrefix(key, args.Key) {
			found = true
			delete(k.Data, key)
		}
	}

	if !found {
		return nil, nil, errors.New("key not found")
	}
	return nil, nil, nil
}

type UpdateArgs struct {
	Key   string        `json:"key"`
	Value string        `json:"value"`
	Index uint64        `json:"index"`
	TTL   time.Duration `json:"ttl"`
}

func (k *KVP) Update(req *acomm.Request) (interface{}, *url.URL, error) {
	var args UpdateArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	value, ok := k.Data[args.Key]
	if ok && value.Index != args.Index {
		return nil, nil, errors.New("CAS failed")
	}
	k.Data[args.Key] = kv.Value{Data: []byte(args.Value), Index: value.Index + 1}
	return map[string]uint64{"index": k.Data[args.Key].Index}, nil, nil
}

func (k *KVP) Ephemeral(req *acomm.Request) (interface{}, *url.URL, error) {
	_, _, err := k.Update(req)
	return nil, nil, err
}
