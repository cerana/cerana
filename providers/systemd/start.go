package systemd

import (
	"errors"
	"net/url"
	"strings"

	"github.com/coreos/go-systemd/dbus"
	"github.com/mistifyio/mistify/acomm"
)

type StartArgs struct {
	Name string `json:"name"`
	Mode string `json:"mode"`
}

type StartResult struct {
	JobID int `json:"job-id"`
}

func (s *Systemd) Start(req *acomm.Request) (interface{}, *url.URL, error) {
	var args StartArgs
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

	resultChan := make(chan string)
	jobID, err := dconn.StartUnit(args.Name, args.Mode, resultChan)
	if err != nil {
		if strings.Contains(err.Error(), "No such file or directory") {
			err = errors.New("unit not found")
		}
		return nil, nil, err
	}
	result := <-resultChan

	if result != "done" {
		err = errors.New(result)
	}

	return &StartResult{jobID}, nil, err
}
