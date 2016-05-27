package systemd

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
)

// ListResult is the result of the List handler.
type ListResult struct {
	Units []UnitStatus `json:"units"`
}

// List retuns a list of unit statuses.
func (s *Systemd) List(req *acomm.Request) (interface{}, *url.URL, error) {
	list, err := s.dconn.ListUnits()
	units := make([]UnitStatus, len(list))
	for i, unit := range list {
		unitStatus, err := s.unitStatus(unit)
		if err != nil {
			return nil, nil, err
		}
		units[i] = *unitStatus
	}

	return &ListResult{units}, nil, err
}
