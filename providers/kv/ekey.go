package kv

import (
	"errors"
	"net/url"
	"time"

	"github.com/cerana/cerana/acomm"
)

var eKeys = newEKeyMap()

// EphemeralSetArgs specifies the arguments to the "kv-ephemeral-set" endpoint.
type EphemeralSetArgs struct {
	Key   string        `json:"key"`
	Value string        `json:"value"`
	TTL   time.Duration `json:"ttl"`
}

// EphemeralDestroyArgs specifies the arguments to the "kv-ephemeral-destroy" endpoint.
type EphemeralDestroyArgs struct {
	Key string `json:"key"`
}

func (k *KV) eset(req *acomm.Request) (interface{}, *url.URL, error) {
	args := EphemeralSetArgs{}
	err := req.UnmarshalArgs(&args)
	if err != nil {
		return nil, nil, err
	}
	if args.Key == "" {
		return nil, nil, errors.New("missing arg: key")
	}
	if args.Value == "" {
		return nil, nil, errors.New("missing arg: value")
	}
	if args.TTL == 0 {
		return nil, nil, errors.New("missing arg: ttl")
	}

	eKey := eKeys.Get(args.Key)
	if eKey == nil || eKey.Renew() != nil {
		eKey, err = k.kv.EphemeralKey(args.Key, args.TTL)
		if err != nil {
			return nil, nil, err
		}
	}

	if err = eKey.Set(args.Value); err != nil {
		return nil, nil, err
	}

	eKeys.Add(args.Key, eKey)
	return nil, nil, nil
}

func (k *KV) edestroy(req *acomm.Request) (interface{}, *url.URL, error) {
	args := EphemeralDestroyArgs{}

	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Key == "" {
		return nil, nil, errors.New("missing arg: key")
	}

	return nil, nil, eKeys.Destroy(args.Key)
}
