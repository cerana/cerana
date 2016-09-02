package kv

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
)

// DeleteArgs specify the arguments to the "kv-delete" endpoint.
type DeleteArgs struct {
	Key       string `json:"key"`
	Recursive bool   `json:"recursive"`
}

// GetArgs specify the arguments to the "kv-get" endpoint.
type GetArgs struct {
	Key string `json:"key"`
}

// SetArgs specify the arguments to the "kv-set" endpoint.
type SetArgs struct {
	Key  string `json:"key"`
	Data string `json:"string"`
}

func (k *KV) isLeader(req *acomm.Request) (interface{}, *url.URL, error) {
	if k.kvDown() {
		return nil, nil, errors.Wrap(errorKVDown)
	}
	leader, err := k.kv.IsLeader()
	return leader, nil, err
}
func (k *KV) delete(req *acomm.Request) (interface{}, *url.URL, error) {
	args := DeleteArgs{}

	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if !args.Recursive && args.Key == "" {
		return nil, nil, errors.Newv("missing arg: key", map[string]interface{}{"args": args})
	}

	if k.kvDown() {
		return nil, nil, errors.Wrap(errorKVDown)
	}
	return nil, nil, k.kv.Delete(args.Key, args.Recursive)
}

func (k *KV) get(req *acomm.Request) (interface{}, *url.URL, error) {
	args := GetArgs{}

	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Key == "" {
		return nil, nil, errors.Newv("missing arg: key", map[string]interface{}{"args": args})
	}

	if k.kvDown() {
		return nil, nil, errors.Wrap(errorKVDown)
	}
	kvp, err := k.kv.Get(args.Key)
	if err != nil {
		return nil, nil, err
	}

	return kvp, nil, nil
}

func (k *KV) getAll(req *acomm.Request) (interface{}, *url.URL, error) {
	args := GetArgs{}

	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Key == "" {
		return nil, nil, errors.Newv("missing arg: key", map[string]interface{}{"args": args})
	}

	if k.kvDown() {
		return nil, nil, errors.Wrap(errorKVDown)
	}
	kvps, err := k.kv.GetAll(args.Key)
	if err != nil {
		return nil, nil, err
	}

	return kvps, nil, nil
}

func (k *KV) keys(req *acomm.Request) (interface{}, *url.URL, error) {
	args := GetArgs{}

	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Key == "" {
		return nil, nil, errors.Newv("missing arg: key", map[string]interface{}{"args": args})
	}

	if k.kvDown() {
		return nil, nil, errors.Wrap(errorKVDown)
	}
	keys, err := k.kv.Keys(args.Key)
	if err != nil {
		return nil, nil, err
	}

	return keys, nil, nil
}

func (k *KV) set(req *acomm.Request) (interface{}, *url.URL, error) {
	args := SetArgs{}

	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Key == "" {
		return nil, nil, errors.Newv("missing arg: key", map[string]interface{}{"args": args})
	}
	if args.Data == "" {
		return nil, nil, errors.Newv("missing arg: data", map[string]interface{}{"args": args})
	}

	if k.kvDown() {
		return nil, nil, errors.Wrap(errorKVDown)
	}
	return nil, nil, k.kv.Set(args.Key, args.Data)
}
