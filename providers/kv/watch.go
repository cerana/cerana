package kv

import (
	"encoding/json"
	"errors"
	"io"
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/kv"
)

var watches = newChanMap()

// WatchArgs specify the arguments to the "kv-watch" endpoint.
type WatchArgs struct {
	Prefix string `json:"prefix"`
	Index  uint64 `json:"index"`
}

// Event specifies structure describing events that took place on watched prefixes.
type Event struct {
	kv.Event
	Error error
}

func makeEventReader(events chan kv.Event, errs chan error) io.ReadCloser {
	r, w := io.Pipe()

	go func() {
		defer func() { _ = r.Close() }()
		defer func() { _ = w.Close() }()

		var event Event
		for {
			select {
			case ev, ok := <-events:
				if !ok {
					return
				}
				event = Event{Event: ev}
			case err, ok := <-errs:
				if !ok {
					return
				}
				event = Event{Error: err}
			}

			data, err := json.Marshal(event)
			if err != nil {
				return
			}
			n, err := w.Write(data)
			if err != nil {
				return
			}
			if n != len(data) {
				return
			}
		}
	}()

	return r
}

func (k *KV) watch(req *acomm.Request) (interface{}, *url.URL, error) {
	args := WatchArgs{}

	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Prefix == "" {
		return nil, nil, errors.New("missing arg: prefix")
	}

	stop := make(chan struct{})
	events, errs, err := k.kv.Watch(args.Prefix, args.Index, stop)
	if err != nil {
		return nil, nil, err
	}

	reader := makeEventReader(events, errs)
	addr, err := k.tracker.NewStreamUnix(k.config.StreamDir("kv-watch"), reader)
	if err != nil {
		return nil, nil, err
	}

	cookie, err := watches.Add(stop)
	if err != nil {
		close(stop)
		return nil, nil, err
	}

	return Cookie{Cookie: uint64(cookie)}, addr, nil
}

func (k *KV) stop(req *acomm.Request) (interface{}, *url.URL, error) {
	args := Cookie{}

	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Cookie == 0 {
		return nil, nil, errors.New("missing arg: cookie")
	}

	ch, err := watches.Get(args.Cookie)
	if err != nil {
		return nil, nil, err
	}

	close(ch)
	return nil, nil, nil
}
