package health

import (
	"net/url"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/providers/systemd"
)

// UptimeArgs are arguments for the uptime health check.
type UptimeArgs struct {
	Name      string        `json:"name"`
	MinUptime time.Duration `json:"minUptime"`
}

// Uptime checks a process's uptime against a minimum value.
func (h *Health) Uptime(req *acomm.Request) (interface{}, *url.URL, error) {
	var args UptimeArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Name == "" {
		return nil, nil, errors.Newv("missing arg: name", map[string]interface{}{"args": args, "missing": "name"})
	}

	unitStatus, err := h.getUnitStatus(args.Name)
	if err != nil {
		return nil, nil, err
	}

	if unitStatus.Uptime < args.MinUptime {
		return nil, nil, errors.Newv("uptime less than expected", map[string]interface{}{"minUptime": args.MinUptime, "uptime": unitStatus.Uptime})
	}

	return nil, nil, nil
}

func (h *Health) getUnitStatus(name string) (*systemd.UnitStatus, error) {
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
		return nil, err
	}
	if err := h.tracker.TrackRequest(req, h.config.RequestTimeout()); err != nil {
		return nil, err
	}
	if err := acomm.Send(h.config.CoordinatorURL(), req); err != nil {
		return nil, err
	}

	resp := <-doneChan
	if resp.Error != nil {
		return nil, errors.ResetStack(resp.Error)
	}

	var result systemd.GetResult
	if err := resp.UnmarshalResult(&result); err != nil {
		return nil, err
	}
	return &result.Unit, nil
}
