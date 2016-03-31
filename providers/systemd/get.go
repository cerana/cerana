package systemd

import (
	"errors"
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/coreos/go-systemd/dbus"
)

// GetArgs are args for the Get handler
type GetArgs struct {
	Name string `json:"name"`
}

// GetResult is the result of the ListUnits handler.
type GetResult struct {
	Unit dbus.UnitStatus `json:"unit"`
}

// Get retuns a list of unit statuses.
func (s *Systemd) Get(req *acomm.Request) (interface{}, *url.URL, error) {
	var args GetArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}

	list, err := s.dconn.ListUnits()
	if err != nil {
		return nil, nil, err
	}

	// Try to find the requested unit in the list
	var res *GetResult
	err = errors.New("unit not found")
	for _, unit := range list {
		if unit.Name == args.Name {
			err = nil
			res = &GetResult{unit}
			break
		}
	}

	return res, nil, err
}
