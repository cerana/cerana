package systemd

import (
	"net/url"

	"github.com/coreos/go-systemd/dbus"
	"github.com/mistifyio/mistify/acomm"
)

// ListUnitsResult is the result of the ListUnits handler.
type ListUnitsResult struct {
	Units []dbus.UnitStatus `json:"units"`
}

// ListUnits retuns a list of unit statuses.
func (s *Systemd) ListUnits(req *acomm.Request) (interface{}, *url.URL, error) {
	dconn, err := dbus.New()
	if err != nil {
		return nil, nil, err
	}
	defer dconn.Close()

	list, err := dconn.ListUnits()
	return &ListUnitsResult{list}, nil, err
}
