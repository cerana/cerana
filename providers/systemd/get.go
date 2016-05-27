package systemd

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/coreos/go-systemd/dbus"
)

type UnitStatus struct {
	dbus.UnitStatus
	Uptime time.Duration
}

// GetArgs are args for the Get handler
type GetArgs struct {
	Name string `json:"name"`
}

// GetResult is the result of the ListUnits handler.
type GetResult struct {
	Unit UnitStatus `json:"unit"`
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
			unitStatus, err := s.unitStatus(unit)
			if err != nil {
				break
			}
			res = &GetResult{*unitStatus}
			break
		}
	}

	return res, nil, err
}

// unitStatus converts a dbus.UnitStatus to a UnitStatus.
func (s *Systemd) unitStatus(unit dbus.UnitStatus) (*UnitStatus, error) {
	unitStatus := &UnitStatus{UnitStatus: unit}

	if unit.ActiveState == "active" {
		prop, err := s.dconn.GetUnitProperty(unit.Name, "ActiveEnterTimestamp")
		if err != nil {
			return nil, err
		}
		activeEnter := time.Unix(int64(prop.Value.Value().(uint64))/int64(time.Second/time.Microsecond), 0)
		unitStatus.Uptime = time.Now().Sub(activeEnter)
		fmt.Println(unit.Name, prop.Value.Value(), activeEnter.Unix(), time.Now().Unix(), unitStatus.Uptime.Seconds())
	}

	return unitStatus, nil
}
