package systemd

import (
	"errors"
	"net/url"

	"github.com/coreos/go-systemd/dbus"
	"github.com/mistifyio/mistify/acomm"
)

// EnableArgs are arguments for the disable handler.
type EnableArgs struct {
	Filepath string `json:"filepath"`
	Runtime  bool   `json:"runtime"`
	Force    bool   `json:"force"`
}

// Enable disables a service.
func (s *Systemd) Enable(req *acomm.Request) (interface{}, *url.URL, error) {
	var args EnableArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Filepath == "" {
		return nil, nil, errors.New("missing arg: filepath")
	}

	dconn, err := dbus.New()
	if err != nil {
		return nil, nil, err
	}
	defer dconn.Close()

	_, _, err = dconn.EnableUnitFiles([]string{args.Filepath}, args.Runtime, args.Force)
	return nil, nil, err
}
