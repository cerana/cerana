package service

import (
	"errors"
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/systemd"
)

type RemoveArgs struct {
	ID       string `json:"id"`
	BundleID uint64 `json:"bundleID"`
}

func (p *Provider) Remove(req *acomm.Request) (interface{}, *url.URL, error) {
	var args RemoveArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == "" {
		return nil, nil, errors.New("missing arg: id")
	}
	if args.BundleID == 0 {
		return nil, nil, errors.New("missing arg: bundleID")
	}

	name := serviceName(args.BundleID, args.ID)
	requests, err := p.prepareRemoveRequests(name)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, p.executeRequests(requests)
}

func (p *Provider) prepareRemoveRequests(name string) ([]*acomm.Request, error) {
	requests := make([]*acomm.Request, 0, 3)
	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:         "systemd-stop",
		ResponseHook: p.tracker.URL(),
		Args: systemd.ActionArgs{
			Name: name,
			Mode: systemd.ModeFail,
		},
	})
	if err != nil {
		return nil, err
	}
	requests = append(requests, req)

	req, err = acomm.NewRequest(acomm.RequestOptions{
		Task:         "systemd-disable",
		ResponseHook: p.tracker.URL(),
		Args: systemd.DisableArgs{
			Name: name,
		},
	})
	if err != nil {
		return nil, err
	}
	requests = append(requests, req)

	req, err = acomm.NewRequest(acomm.RequestOptions{
		Task:         "systemd-remove",
		ResponseHook: p.tracker.URL(),
		Args: systemd.ActionArgs{
			Name: name,
		},
	})
	if err != nil {
		return nil, err
	}
	requests = append(requests, req)
	return requests, nil
}
