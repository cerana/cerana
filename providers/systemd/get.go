package systemd

import (
	"errors"
	"net/url"

	"github.com/coreos/go-systemd/dbus"
	"github.com/mistifyio/mistify/acomm"
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

	dconn, err := dbus.New()
	if err != nil {
		return nil, nil, err
	}
	defer dconn.Close()

	list, err := dconn.ListUnits()
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
