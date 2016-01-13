package simple

import (
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/acomm"
)

type Simple struct{}

type SystemStatusArgs struct {
	GuestID string
}
type SystemStatusResult struct {
	CPU  []*CPUInfo
	Disk []*DiskInfo
}

type CPUInfoArgs struct {
	GuestID string
}
type CPUInfoResult []*CPUInfo
type CPUInfo struct {
	Processor int
	MHz       int
}

type DiskInfoArgs struct {
	GuestID string
}
type DiskInfoResult []*DiskInfo
type DiskInfo struct {
	Device string
	Size   int64
}

func SendToController(req *acomm.Request) error {
	return nil
}

func (s *Simple) SystemStatus(argI interface{}) (interface{}, error) {
	args, ok := argI.(SystemStatusArgs)
	if !ok {
		return nil, errors.New("invalid arguments")
	}

	// Prepare multiple requests
	multiRequest := NewMultiRequest(nil)

	cpuReq, err := acomm.NewRequest("CPUInfo", "", &CPUInfoArgs{GuestID: args.GuestID}, nil, nil)
	if err != nil {
		return nil, err
	}
	diskReq, err := acomm.NewRequest("DiskInfo", "", &DiskInfoArgs{GuestID: args.GuestID}, nil, nil)
	if err != nil {
		return nil, err
	}

	requests := map[string]*acomm.Request{
		"CPUInfo":  cpuReq,
		"DiskInfo": diskReq,
	}

	for name, req := range requests {
		multiRequest.AddRequest(name, req)
		if err := SendToController(req); err != nil {
			multiRequest.RemoveRequest(req.ID)
		}
	}

	// Wait for the results
	results, errs := multiRequest.Results()
	if errs != nil {
		err := errors.New("one or more status requests failed")
		log.WithFields(log.Fields{
			"requests": requests,
			"errors":   errs,
		}).Error(err)
		return nil, err
	}

	result := &SystemStatusResult{
		CPU:  results["CPUInfo"].(CPUInfoResult),
		Disk: results["DiskInfo"].(DiskInfoResult),
	}
	return result, nil
}

func (s *Simple) CPUInfo(argI interface{}) (interface{}, error) {
	args, ok := argI.(CPUInfoArgs)
	if !ok {
		return nil, errors.New("invalid arguments")
	}

	if args.GuestID == "" {
		return nil, errors.New("missing guest id")
	}

	result := &CPUInfoResult{
		&CPUInfo{
			Processor: 0,
			MHz:       2600,
		},
		&CPUInfo{
			Processor: 1,
			MHz:       2600,
		},
	}
	return result, nil
}

func (s *Simple) DiskInfo(argI interface{}) (interface{}, error) {
	args, ok := argI.(DiskInfoArgs)
	if !ok {
		return nil, errors.New("invalid arguments")
	}

	if args.GuestID == "" {
		return nil, errors.New("missing guest id")
	}

	result := &DiskInfoResult{
		&DiskInfo{
			Device: "vda1",
			Size:   10 * (1024 * 1024 * 1024), // 10 GB in bytes
		},
	}

	return result, nil
}
