package health

import (
	"errors"
	"net/url"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/systemd"
)

type UptimeArgs struct {
	Name      string        `json:"name"`
	MinUptime time.Duration `json:"minUptime"`
}

func (h *Health) Uptime(req *acomm.Request) (interface{}, *url.URL, error) {
	var args UptimeArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Name == "" {
		return nil, nil, errors.New("missing arg: path")
	}

	unitStatus, err := h.getUnitStatus(args.Name)
	if err != nil {
		return nil, nil, err
	}

	if unitStatus.Uptime < args.MinUptime {
		return nil, nil, errors.New("uptime less than expected")
	}

	return nil, nil, nil
}

func (h *Health) getUnitStatus(name string) (systemd.UnitStatus, error) {
	var unitStatus systemd.UnitStatus
	doneChan := make(chan *acomm.Response, 1)
	rh := func(_ *acomm.Request, resp *acomm.Response) {
		doneChan <- resp
	}
	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:           "systemd-get",
		ResponseHook:   h.tracker.URL(),
		Args:           &systemd.GetArgs{Name: name},
		SuccessHandler: rh,
		ErrorHandler:   rh,
	})

	if err != nil {
		return unitStatus, err
	}
	if err := h.tracker.TrackRequest(req, h.config.RequestTimeout()); err != nil {
		return unitStatus, err
	}
	if err := acomm.Send(h.config.CoordinatorURL(), req); err != nil {
		return unitStatus, err
	}

	resp := <-doneChan
	if resp.Error != nil {
		return unitStatus, resp.Error
	}

	return unitStatus, resp.UnmarshalResult(&unitStatus)
}
