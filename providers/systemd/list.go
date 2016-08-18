package systemd

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
)

// ListResult is the result of the List handler.
type ListResult struct {
	Units []UnitStatus `json:"units"`
}

// List retuns a list of unit statuses.
func (s *Systemd) List(req *acomm.Request) (interface{}, *url.URL, error) {
	list, err := s.dconn.ListUnits()
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}

	units := make([]UnitStatus, len(list))
	for i, unit := range list {
		var unitStatus *UnitStatus
		unitStatus, err = s.unitStatus(unit)
		if err != nil {
			return nil, nil, err
		}
		units[i] = *unitStatus
	}

	return &ListResult{units}, nil, nil
}
