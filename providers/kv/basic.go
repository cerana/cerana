package kv

import (
	"errors"
	"net/url"

	"github.com/cerana/cerana/acomm"
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

func (k *KV) delete(req *acomm.Request) (interface{}, *url.URL, error) {
	args := DeleteArgs{}

	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Key == "" {
		return nil, nil, errors.New("missing arg: key")
	}

	return nil, nil, k.kv.Delete(args.Key, args.Recursive)
}

func (k *KV) get(req *acomm.Request) (interface{}, *url.URL, error) {
	args := GetArgs{}

	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Key == "" {
		return nil, nil, errors.New("missing arg: key")
	}

	kvp, err := k.kv.Get(args.Key)
	if err != nil {
		return nil, nil, err
	}

	return kvp.Data, nil, nil
}

func (k *KV) getAll(req *acomm.Request) (interface{}, *url.URL, error) {
	args := GetArgs{}

	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Key == "" {
		return nil, nil, errors.New("missing arg: key")
	}

	kvp, err := k.kv.GetAll(args.Key)
	if err != nil {
		return nil, nil, err
	}

	return kvp, nil, nil
}

func (k *KV) keys(req *acomm.Request) (interface{}, *url.URL, error) {
	args := GetArgs{}

	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Key == "" {
		return nil, nil, errors.New("missing arg: key")
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
		return nil, nil, errors.New("missing arg: key")
	}
	if args.Data == "" {
		return nil, nil, errors.New("missing arg: data")
	}

	return nil, nil, k.kv.Set(args.Key, args.Data)
}
