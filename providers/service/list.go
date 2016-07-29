package service

import (
	"fmt"
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
	defer close(ch)

	rh := func(_ *acomm.Request, resp *acomm.Response) {
		ch <- resp
	}
	fmt.Println("List: creating request")
	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:           "systemd-list",
		ResponseHook:   p.tracker.URL(),
		SuccessHandler: rh,
		ErrorHandler:   rh,
	})
	if err != nil {
		fmt.Println("List: create request error")
		return nil, nil, err
	}
	fmt.Println("List: tracking request")
	if err := p.tracker.TrackRequest(req, 0); err != nil {
		return nil, nil, err
	}
	fmt.Println("List: request tracked")
	if err := acomm.Send(p.config.CoordinatorURL(), req); err != nil {
		p.tracker.RemoveRequest(req)
		return nil, nil, err
	}
	fmt.Println("List: request sent")

	resp := <-ch
	if resp.Error != nil {
		fmt.Println("List: response error")
		return nil, nil, resp.Error
	}

	fmt.Println("List: unmarshalling response")
	var listResult systemd.ListResult
	if err := resp.UnmarshalResult(&listResult); err != nil {
		return nil, nil, err
	}
	systemdUnits := listResult.Units

	fmt.Println("List: making services array from response")

	services := make([]Service, 0, len(systemdUnits))
	for _, unit := range systemdUnits {
		service, err := systemdUnitToService(unit)
		if err != nil {
			continue
		}
		services = append(services, *service)
	}

	fmt.Println("List: returning result")
	return ListResult{services}, nil, nil
}
