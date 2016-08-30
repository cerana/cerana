package service

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/providers/systemd"
)

// Service is information about a service.
type Service struct {
	ID          string            `json:"id"`
	BundleID    uint64            `json:"bundleID"`
	Description string            `json:"description"`
	Uptime      time.Duration     `json:"uptime"`
	ActiveState string            `json:"activeState"`
	Cmd         []string          `json:"cmd"`
	UID         uint64            `json:"uid"`
	GID         uint64            `json:"gid"`
	Env         map[string]string `json:"env"`
}

// GetArgs are args for retrieving a service.
type GetArgs struct {
	ID       string `json:"id"`
	BundleID uint64 `json:"bundleID"`
}

// GetResult is the result of a Get.
type GetResult struct {
	Service Service `json:"service"`
}

// Get retrieves a service.
func (p *Provider) Get(req *acomm.Request) (interface{}, *url.URL, error) {
	var args GetArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == "" {
		return nil, nil, errors.Newv("missing arg: id", map[string]interface{}{"args": args, "missing": "id"})
	}
	if args.BundleID == 0 {
		return nil, nil, errors.Newv("missing arg: bundleID", map[string]interface{}{"args": args, "missing": "bundleID"})
	}

	service, err := p.getService(args.BundleID, args.ID)
	if err != nil {
		return nil, nil, err
	}

	return GetResult{*service}, nil, nil
}

func (p *Provider) getService(bundleID uint64, id string) (*Service, error) {
	name := serviceName(bundleID, id)

	// Request
	ch := make(chan *acomm.Response, 1)
	rh := func(_ *acomm.Request, resp *acomm.Response) {
		ch <- resp
	}
	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:         "systemd-get",
		ResponseHook: p.tracker.URL(),
		Args: systemd.GetArgs{
			Name: name,
		},
		SuccessHandler: rh,
		ErrorHandler:   rh,
	})
	if err != nil {
		return nil, errors.Wrap(err)
	}
	if err := p.tracker.TrackRequest(req, 0); err != nil {
		return nil, err
	}
	if err := acomm.Send(p.config.CoordinatorURL(), req); err != nil {
		p.tracker.RemoveRequest(req)
		return nil, err
	}

	resp := <-ch
	if resp.Error != nil {
		return nil, errors.ResetStack(resp.Error)
	}

	var getResult systemd.GetResult
	if err := resp.UnmarshalResult(&getResult); err != nil {
		return nil, err
	}

	return systemdUnitToService(getResult.Unit)
}

func serviceName(bundleID uint64, serviceID string) string {
	return fmt.Sprintf("%d:%s.service", bundleID, serviceID)
}

func systemdUnitToService(systemdUnit systemd.UnitStatus) (*Service, error) {
	r, _ := regexp.Compile(`^(\d+):(.+)\.service$`)

	idParts := r.FindStringSubmatch(systemdUnit.Name)
	if len(idParts) != 3 {
		return nil, errors.Newv("service name not correct format", map[string]interface{}{"serviceName": systemdUnit.Name})
	}

	// only a valid uint would make it this far
	bundleID, _ := strconv.ParseUint(idParts[1], 10, 64)

	env := make(map[string]string)
	if systemdUnit.UnitTypeProperties["Environment"] != nil {
		for _, kv := range systemdUnit.UnitTypeProperties["Environment"].([]interface{}) {
			parts := strings.SplitN(kv.(string), "=", 2)
			env[parts[0]] = parts[1]
		}
	}

	execStartInterface, ok := systemdUnit.UnitTypeProperties["ExecStart"]
	var execStart []string
	if ok {
		// [["cmd",["cmd","param","param"],0,...]]
		tmp := execStartInterface.([]interface{})[0].([]interface{})[1].([]interface{})
		execStart = make([]string, len(tmp))
		for i, v := range tmp {
			execStart[i] = v.(string)
		}
	}

	var err error

	uidS, ok := systemdUnit.UnitTypeProperties["User"].(string)
	uid := uint64(0)
	if ok && uidS != "" {
		uid, err = strconv.ParseUint(uidS, 10, 64)
		if err != nil {
			return nil, errors.Wrapv(err, map[string]interface{}{"uid": uidS})
		}
	}

	gidS, ok := systemdUnit.UnitTypeProperties["Group"].(string)
	gid := uint64(0)
	if ok && gidS != "" {
		gid, err = strconv.ParseUint(gidS, 10, 64)
		if err != nil {
			return nil, errors.Wrapv(err, map[string]interface{}{"gid": gid})
		}
	}

	descriptionInterface, ok := systemdUnit.UnitProperties["Description"]
	description := ""
	if ok {
		description = descriptionInterface.(string)
	}

	service := &Service{
		ID:          idParts[2],
		BundleID:    bundleID,
		Description: description,
		Uptime:      systemdUnit.Uptime,
		ActiveState: systemdUnit.ActiveState,
		Cmd:         execStart,
		// TODO: Possibly extract User and Group from Cmd
		UID: uid,
		GID: gid,
		Env: env,
	}

	return service, nil
}
