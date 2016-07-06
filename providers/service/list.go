package service

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/systemd"
)

// ListResult is the result of the List handler.
type ListResult struct {
	Services []Service
}

// List returns a list of Services and information about each.
func (p *Provider) List(req *acomm.Request) (interface{}, *url.URL, error) {
	ch := make(chan *acomm.Response, 1)
	rh := func(_ *acomm.Request, resp *acomm.Response) {
		ch <- resp
	}
	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:           "systemd-list",
		ResponseHook:   p.tracker.URL(),
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
	if resp.Error != nil {
		return nil, nil, resp.Error
	}

	var listResult systemd.ListResult
	if err := resp.UnmarshalResult(&listResult); err != nil {
		return nil, nil, err
	}
	systemdUnits := listResult.Units

	services := make([]Service, 0, len(systemdUnits))
	for _, unit := range systemdUnits {
		service, err := systemdUnitToService(unit)
		if err != nil {
			continue
		}
		services = append(services, *service)
	}

	return ListResult{services}, nil, nil
}
