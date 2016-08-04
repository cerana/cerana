package kv

import (
	"errors"
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/kv"
)

// RemoveArgs specifies the arguments to the "kv-remove" endpoint.
type RemoveArgs struct {
	Key   string `json:"key"`
	Index uint64 `json:"index"`
}

// UpdateArgs specifies the arguments to the "kv-update" endpoint.
type UpdateArgs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Index uint64 `json:"index"`
}

// UpdateReturn specifies the return value from the "kv-update" endpoint.
type UpdateReturn struct {
	Index uint64 `json:"index"`
}

func (k *KV) remove(req *acomm.Request) (interface{}, *url.URL, error) {
	args := RemoveArgs{}

	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Key == "" {
		return nil, nil, errors.New("missing arg: key")
	}

	if k.kvDown() {
		return nil, nil, errorKVDown
	}
	return nil, nil, k.kv.Remove(args.Key, args.Index)
}

func (k *KV) update(req *acomm.Request) (interface{}, *url.URL, error) {
	args := UpdateArgs{}

	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Key == "" {
		return nil, nil, errors.New("missing arg: key")
	}
	if args.Value == "" {
		return nil, nil, errors.New("missing arg: value")
	}

	value := kv.Value{
		Data:  []byte(args.Value),
		Index: args.Index,
	}

	if k.kvDown() {
		return nil, nil, errorKVDown
	}
	index, err := k.kv.Update(args.Key, value)
	if err != nil {
		return nil, nil, err
	}

	ret := UpdateReturn{
		Index: index,
	}
	return ret, nil, nil
}
