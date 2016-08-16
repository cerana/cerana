package service

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/providers/systemd"
)

// RestartArgs are arguments for Restart.
type RestartArgs struct {
	ID       string `json:"id"`
	BundleID uint64 `json:"bundleID"`
}

// Restart restarts a service.
func (p *Provider) Restart(req *acomm.Request) (interface{}, *url.URL, error) {
	var args RestartArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == "" {
		return nil, nil, errors.Newv("missing arg: id", map[string]interface{}{"args": args})
	}
	if args.BundleID == 0 {
		return nil, nil, errors.Newv("missing arg: bundleID", map[string]interface{}{"args": args})
	}

	ch := make(chan *acomm.Response, 1)
	rh := func(_ *acomm.Request, resp *acomm.Response) {
		ch <- resp
	}
	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:         "systemd-restart",
		ResponseHook: p.tracker.URL(),
		Args: systemd.ActionArgs{
			Name: serviceName(args.BundleID, args.ID),
			Mode: systemd.ModeFail,
		},
		SuccessHandler: rh,
		ErrorHandler:   rh,
	})
	if err != nil {
		return nil, nil, err
	}
	if err := p.tracker.TrackRequest(req, 0); err != nil {
		return nil, nil, err
	}
	if err := acomm.Send(p.config.CoordinatorURL(), req); err != nil {
		p.tracker.RemoveRequest(req)
		return nil, nil, err
	}

	resp := <-ch
	return nil, nil, errors.Wrap(resp.Error)
}
