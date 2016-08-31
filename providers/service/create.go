package service

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/providers/systemd"
	"github.com/coreos/go-systemd/unit"
)

// CreateArgs contains args for creating or replacing a Service.
type CreateArgs struct {
	ID          string            `json:"id"`
	BundleID    uint64            `json:"bundleID"`
	Dataset     string            `json:"dataset"`
	Description string            `json:"description"`
	Cmd         []string          `json:"cmd"`
	Env         map[string]string `json:"env"`
	Overwrite   bool              `json:"overwrite"`
}

// Create creates (or replaces) and starts (or restarts) a service.
func (p *Provider) Create(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CreateArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	argErrData := map[string]interface{}{"args": args}

	if args.ID == "" {
		argErrData["missing"] = "id"
		return nil, nil, errors.Newv("missing arg: id", argErrData)
	}
	if args.BundleID == 0 {
		argErrData["missing"] = "bundleID"
		return nil, nil, errors.Newv("missing arg: bundleID", argErrData)
	}
	if len(args.Cmd) == 0 {
		argErrData["missing"] = "cmd"
		return nil, nil, errors.Newv("missing arg: cmd", argErrData)
	}
	if args.Dataset == "" {
		argErrData["missing"] = "dataset"
		return nil, nil, errors.Newv("missing arg: dataset", argErrData)
	}

	name := serviceName(args.BundleID, args.ID)
	datasetCloneName := filepath.Join(p.config.DatasetCloneDir(), name)
	daisyEnvParts := make([]string, 0, len(args.Env))
	for key, val := range args.Env {
		daisyEnvParts = append(daisyEnvParts, fmt.Sprintf("%s=%s", key, val))
	}
	envArgs := strings.Join(daisyEnvParts, ",")
	envString := ""
	if envArgs != "" {
		envString = "-e " + envArgs
	}
	execStart := fmt.Sprintf("/run/current-system/sw/bin/daisy %s -r /%s %s", envString, datasetCloneName, strings.Join(args.Cmd, " "))
	unitOptions := []*unit.UnitOption{
		{Section: "Unit", Name: "Description", Value: args.Description},
		// TODO: Does cmd get prepended with daisy?
		{Section: "Service", Name: "ExecStart", Value: execStart},
		{Section: "Service", Name: "Type", Value: "simple"},
		{Section: "Service", Name: "Restart", Value: "always"},
		{Section: "Service", Name: "RestartSec", Value: "3"},
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

	requests, continueChecks, err := p.prepareCreateRequests(name, unitOptions, args.Overwrite)
	if err != nil {
		return nil, nil, err
	}

	if err = p.executeRequests(requests, continueChecks); err != nil {
		return nil, nil, err
	}

	service, err := p.getService(args.BundleID, args.ID)
	if err != nil {
		return nil, nil, err
	}
	return GetResult{*service}, nil, nil
}

func (p *Provider) prepareCreateRequests(name string, unitOptions []*unit.UnitOption, overwrite bool) ([]*acomm.Request, []continueCheck, error) {
	requests := make([]*acomm.Request, 0, 3)
	continueChecks := make([]continueCheck, 0, 3)

	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:         "systemd-create",
		ResponseHook: p.tracker.URL(),
		Args: systemd.CreateArgs{
			Name:        name,
			UnitOptions: unitOptions,
			Overwrite:   overwrite,
		},
	})
	if err != nil {
		return nil, nil, err
	}
	requests = append(requests, req)
	// there's only more work to do if the unit was modified
	continueChecks = append(continueChecks, func(resp *acomm.Response) (bool, error) {
		var result systemd.CreateResult
		if err = resp.UnmarshalResult(&result); err != nil {
			return false, err
		}
		return result.UnitModified, nil
	})

	req, err = acomm.NewRequest(acomm.RequestOptions{
		Task:         "systemd-enable",
		ResponseHook: p.tracker.URL(),
		Args: systemd.EnableArgs{
			Name: name,
		},
	})
	if err != nil {
		return nil, nil, err
	}
	requests = append(requests, req)
	continueChecks = append(continueChecks, nil)

	req, err = acomm.NewRequest(acomm.RequestOptions{
		Task:         "systemd-restart",
		ResponseHook: p.tracker.URL(),
		Args: systemd.ActionArgs{
			Name: name,
			Mode: systemd.ModeFail,
		},
	})
	if err != nil {
		return nil, nil, err
	}
	requests = append(requests, req)
	continueChecks = append(continueChecks, nil)

	return requests, continueChecks, nil
}
