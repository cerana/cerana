package systemd

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
)

// DisableArgs are arguments for the disable handler.
type DisableArgs struct {
	Name    string `json:"name"`
	Runtime bool   `json:"runtime"`
}

// Disable disables a service.
func (s *Systemd) Disable(req *acomm.Request) (interface{}, *url.URL, error) {
	var args DisableArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Name == "" {
		return nil, nil, errors.Newv("missing arg: name", map[string]interface{}{"args": args})
	}

	_, err := s.dconn.DisableUnitFiles([]string{args.Name}, args.Runtime)
	return nil, nil, errors.Wrapv(err, map[string]interface{}{"name": args.Name, "runtime": args.Runtime})
}
