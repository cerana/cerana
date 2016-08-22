package systemd

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
)

// EnableArgs are arguments for the disable handler.
type EnableArgs struct {
	Name    string `json:"name"`
	Runtime bool   `json:"runtime"`
	Force   bool   `json:"force"`
}

// Enable disables a service.
func (s *Systemd) Enable(req *acomm.Request) (interface{}, *url.URL, error) {
	var args EnableArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Name == "" {
		return nil, nil, errors.Newv("missing arg: name", map[string]interface{}{"args": args})
	}

	unitFilePath, err := s.config.UnitFilePath(args.Name)
	if err != nil {
		return nil, nil, err
	}

	_, _, err = s.dconn.EnableUnitFiles([]string{unitFilePath}, args.Runtime, args.Force)
	return nil, nil, errors.Wrapv(err, map[string]interface{}{"path": unitFilePath, "runtime": args.Runtime, "force": args.Force})
}
