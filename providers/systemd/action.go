package systemd

import (
	"net/url"
	"strings"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
)

// ActionArgs are arguments for service running action handlers.
type ActionArgs struct {
	Name string `json:"name"`
	Mode string `json:"mode"`
}

// Start starts an enabled service.
func (s *Systemd) Start(req *acomm.Request) (interface{}, *url.URL, error) {
	return s.systemdAction("start", req)
}

// Stop stops a running service.
func (s *Systemd) Stop(req *acomm.Request) (interface{}, *url.URL, error) {
	return s.systemdAction("stop", req)
}

// Restart restarts a service.
func (s *Systemd) Restart(req *acomm.Request) (interface{}, *url.URL, error) {
	return s.systemdAction("restart", req)
}

// systemdAction is a common method for systemd service running actions.
func (s *Systemd) systemdAction(action string, req *acomm.Request) (interface{}, *url.URL, error) {
	var args ActionArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Name == "" {
		return nil, nil, errors.Newv("missing arg: name", map[string]interface{}{"args": args})
	}

	var actionFn func(string, string, chan<- string) (int, error)
	switch action {
	case "start":
		actionFn = s.dconn.StartUnit
	case "stop":
		actionFn = s.dconn.StopUnit
	case "restart":
		actionFn = s.dconn.RestartUnit
	}

	// Run the action. Ignore jobid since we are waiting for the result; once a
	// job is completed, the jobid is meaningless.
	if _, err := actionFn(args.Name, args.Mode, nil); err != nil {
		if strings.Contains(err.Error(), "No such file or directory") {
			err = errors.Newv("unit not found", map[string]interface{}{"name": args.Name})
		}
		return nil, nil, errors.Wrapv(err, map[string]interface{}{"name": args.Name, "mode": args.Mode, "action": action})
	}

	return nil, nil, nil
}
