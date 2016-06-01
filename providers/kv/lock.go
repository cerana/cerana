package kv

import (
	"errors"
	"net/url"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/kv"
)

var locks = newLockMap()

// LockArgs specifies the arguments to the "kv-lock" endpoint.
type LockArgs struct {
	Key string        `json:"key"`
	TTL time.Duration `json:"ttl"`
}

func (k *KV) lock(req *acomm.Request) (interface{}, *url.URL, error) {
	args := LockArgs{}

	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Key == "" {
		return nil, nil, errors.New("missing arg: key")
	}
	if args.TTL == 0 {
		return nil, nil, errors.New("missing arg: ttl")
	}

	lock, err := k.kv.Lock(args.Key, args.TTL)
	if err != nil {
		return nil, nil, err
	}

	cookie, err := locks.Add(lock)
	if err != nil {
		_ = lock.Unlock()
		return nil, nil, err
	}

	return Cookie{Cookie: cookie}, nil, nil
}

func (k *KV) renew(req *acomm.Request) (interface{}, *url.URL, error) {
	args := Cookie{}

	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Cookie == 0 {
		return nil, nil, errors.New("missing arg: cookie")
	}

	iface, err := locks.Peek(args.Cookie)
	if err != nil {
		return nil, nil, err
	}

	switch EorL := iface.(type) {
	case kv.Lock:
		err = EorL.Renew()
	case kv.EphemeralKey:
		err = EorL.Renew()
	}

	return nil, nil, err
}

func (k *KV) unlock(req *acomm.Request) (interface{}, *url.URL, error) {
	args := Cookie{}

	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Cookie == 0 {
		return nil, nil, errors.New("missing arg: cookie")
	}

	iface, err := locks.Get(args.Cookie)
	if err != nil {
		return nil, nil, err
	}

	switch EorL := iface.(type) {
	case kv.Lock:
		err = EorL.Unlock()
	case kv.EphemeralKey:
		err = EorL.Destroy()
	}

	return nil, nil, err
}
