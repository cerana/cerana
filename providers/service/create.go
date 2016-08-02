package service

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/systemd"
	"github.com/coreos/go-systemd/unit"
)

// CreateArgs contains args for creating a new Service.
type CreateArgs struct {
	ID          string            `json:"id"`
	BundleID    uint64            `json:"bundleID"`
	Dataset     string            `json:"dataset"`
	Description string            `json:"description"`
	Cmd         []string          `json:"cmd"`
	Env         map[string]string `json:"env"`
}

// Create creates and starts a new service.
func (p *Provider) Create(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CreateArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == "" {
		return nil, nil, errors.New("missing arg: id")
	}
	if args.BundleID == 0 {
		return nil, nil, errors.New("missing arg: bundleID")
	}
	if len(args.Cmd) == 0 {
		return nil, nil, errors.New("missing arg: cmd")
	}
	if args.Dataset == "" {
		return nil, nil, errors.New("missing arg: dataset")
	}

	name := serviceName(args.BundleID, args.ID)
	datasetCloneName := filepath.Join(p.config.DatasetCloneDir(), name)
	unitOptions := []*unit.UnitOption{
		{Section: "Unit", Name: "Description", Value: args.Description},
		// TODO: Does cmd get prepended with daisy?
		{Section: "Service", Name: "ExecStart", Value: strings.Join(args.Cmd, " ")},
		{Section: "Service", Name: "Type", Value: "forking"},
		{Section: "Install", Name: "WantedBy", Value: "cerana.target"},

		{Section: "Service", Name: "ExecStartPre", Value: p.config.RollbackCloneCmd()},
		{Section: "Service", Name: " ExecStartPre", Value: fmt.Sprintf("/run/current-system/sw/bin/mkdir -p /%s/etc", datasetCloneName)},
		{Section: "Service", Name: " ExecStartPre", Value: fmt.Sprintf("/run/current-system/sw/bin/touch /%s/etc/machine-id", datasetCloneName)},
		{Section: "Service", Name: "Environment", Value: "_CERANA_CLONE_SOURCE=" + args.Dataset},
		{Section: "Service", Name: "Environment", Value: "_CERANA_CLONE_DESTINATION=" + datasetCloneName},
	}
	// TODO: Add User= and Group= if not part of daisy
	for key, val := range args.Env {
		// do not allow custom overrides of the internal cerana env variables
		if strings.HasPrefix(key, "_CERANA_") {
			continue
		}
		unitOptions = append(unitOptions, &unit.UnitOption{
			Section: "Service",
			Name:    "Environment",
			Value:   fmt.Sprintf("%s=%s", key, val),
		})
	}

	requests, err := p.prepareCreateRequests(name, unitOptions)
	if err != nil {
		return nil, nil, err
	}

	if err = p.executeRequests(requests); err != nil {
		return nil, nil, err
	}

	service, err := p.getService(args.BundleID, args.ID)
	if err != nil {
		return nil, nil, err
	}
	return GetResult{*service}, nil, nil
}

func (p *Provider) prepareCreateRequests(name string, unitOptions []*unit.UnitOption) ([]*acomm.Request, error) {
	requests := make([]*acomm.Request, 0, 3)
	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:         "systemd-create",
		ResponseHook: p.tracker.URL(),
		Args: systemd.CreateArgs{
			Name:        name,
			UnitOptions: unitOptions,
		},
	})
	if err != nil {
		return nil, err
	}
	requests = append(requests, req)

	req, err = acomm.NewRequest(acomm.RequestOptions{
		Task:         "systemd-enable",
		ResponseHook: p.tracker.URL(),
		Args: systemd.EnableArgs{
			Name: name,
		},
	})
	if err != nil {
		return nil, err
	}
	requests = append(requests, req)

	req, err = acomm.NewRequest(acomm.RequestOptions{
		Task:         "systemd-start",
		ResponseHook: p.tracker.URL(),
		Args: systemd.ActionArgs{
			Name: name,
			Mode: systemd.ModeFail,
		},
	})
	if err != nil {
		return nil, err
	}
	requests = append(requests, req)
	return requests, nil
}
