package systemd

import (
	"errors"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/coreos/go-systemd/dbus"
)

// UnitStatus contains information about a systemd unit.
type UnitStatus struct {
	dbus.UnitStatus
	Uptime             time.Duration          `json:"uptime"`
	UnitProperties     map[string]interface{} `json:"unitProperties"`
	UnitTypeProperties map[string]interface{} `json:"unitTypeProperties"`
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
			var unitStatus *UnitStatus
			unitStatus, err = s.unitStatus(unit)
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

	unitProps, err := s.dconn.GetUnitProperties(unit.Name)
	if err != nil {
		return nil, err
	}
	unitStatus.UnitProperties = unitProps

	unitType := strings.Title(strings.TrimLeft(filepath.Ext(unit.Name), "."))
	unitTypeProps, err := s.dconn.GetUnitTypeProperties(unit.Name, unitType)
	if err != nil && !strings.Contains(err.Error(), "Unknown interface") {
		return nil, err
	}
	unitStatus.UnitTypeProperties = unitTypeProps

	if unitStatus.ActiveState == "active" {
		activeEnterDur := time.Duration(unitStatus.UnitProperties["ActiveEnterTimestamp"].(uint64)) * time.Microsecond
		activeEnterTs := time.Unix(int64(activeEnterDur.Seconds()), 0)
		unitStatus.Uptime = time.Now().Sub(activeEnterTs)
	}

	return unitStatus, nil
}
