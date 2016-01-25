package simple

import (
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/acomm"
)

// Simple is a simple provider implementation.
type Simple struct {
	config  *Config
	tracker *acomm.Tracker
}

// SystemStatusArgs are arguments for the SystemStatus handler.
type SystemStatusArgs struct {
	GuestID string `json:"guest_id"`
}

// SystemStatusResult is the result data for the SystemStatus handler.
type SystemStatusResult struct {
	CPUs  []*CPUInfo  `json:"cpus"`
	Disks []*DiskInfo `json:"disks"`
}

// CPUInfoArgs are arguments for the CPUInfo handler.
type CPUInfoArgs struct {
	GuestID string `json:"guest_id"`
}

// CPUInfoResult is the result data for the CPUInfo handler.
type CPUInfoResult []*CPUInfo

// CPUInfo is information on a particular CPU.
type CPUInfo struct {
	Processor int `json:"processor"`
	MHz       int `json:"mhz"`
}

// DiskInfoArgs are arguments for the DiskInfo handler.
type DiskInfoArgs struct {
	GuestID string `json:"guest_id"`
}

// DiskInfoResult is the result data for the DiskInfo handler.
type DiskInfoResult []*DiskInfo

// DiskInfo is information on a particular disk.
type DiskInfo struct {
	Device string
	Size   int64
}

// NewSimple creates a new instance of Simple.
func NewSimple(config *Config, tracker *acomm.Tracker) *Simple {
	return &Simple{
		config:  config,
		tracker: tracker,
	}
}

// RegisterTasks registers all of Simple's task handlers with the server.
func (s *Simple) RegisterTasks(server *Server) {
	server.RegisterTask("SystemStatus", s.SystemStatus)
	server.RegisterTask("CPUInfo", s.CPUInfo)
	server.RegisterTask("DiskInfo", s.DiskInfo)
}

// SystemStatus is a task handler to retrieve info look up and return system
// information. It depends on and makes requests for several other tasks.
func (s *Simple) SystemStatus(req *acomm.Request) (interface{}, error) {
	var args SystemStatusArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, err
	}
	if args.GuestID == "" {
		return nil, errors.New("missing guest_id")
	}

	// Prepare multiple requests
	multiRequest := NewMultiRequest(s.tracker)

	cpuReq, err := acomm.NewRequest("CPUInfo", s.tracker.URL().String(), &CPUInfoArgs{GuestID: args.GuestID}, nil, nil)
	if err != nil {
		return nil, err
	}
	diskReq, err := acomm.NewRequest("DiskInfo", s.tracker.URL().String(), &DiskInfoArgs{GuestID: args.GuestID}, nil, nil)
	if err != nil {
		return nil, err
	}

	requests := map[string]*acomm.Request{
		"CPUInfo":  cpuReq,
		"DiskInfo": diskReq,
	}

	for name, req := range requests {
		if err := multiRequest.AddRequest(name, req); err != nil {
			continue
		}
		if err := acomm.Send(s.config.CoordinatorURL(), req); err != nil {
			multiRequest.RemoveRequest(req)
			continue
		}
	}

	// Wait for the results
	responses := multiRequest.Responses()
	result := &SystemStatusResult{}

	if resp, ok := responses["CPUInfo"]; ok {
		if err := resp.UnmarshalResult(&(result.CPUs)); err != nil {
			log.WithFields(log.Fields{
				"name":  "CPUInfo",
				"resp":  resp,
				"error": err,
			}).Error("failed to unarshal result")
		}
	}

	if resp, ok := responses["DiskInfo"]; ok {
		if err := resp.UnmarshalResult(&(result.Disks)); err != nil {
			log.WithFields(log.Fields{
				"name":  "DiskInfo",
				"resp":  resp,
				"error": err,
			}).Error("failed to unarshal result")
		}
	}

	return result, nil
}

// CPUInfo is a task handler to retrieve information about CPUs.
func (s *Simple) CPUInfo(req *acomm.Request) (interface{}, error) {
	var args CPUInfoArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, err
	}
	if args.GuestID == "" {
		return nil, errors.New("missing guest_id")
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

// DiskInfo is a task handler to retrieve information about disks.
func (s *Simple) DiskInfo(req *acomm.Request) (interface{}, error) {
	var args CPUInfoArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, err
	}
	if args.GuestID == "" {
		return nil, errors.New("missing guest_id")
	}

	result := &DiskInfoResult{
		&DiskInfo{
			Device: "vda1",
			Size:   10 * (1024 * 1024 * 1024), // 10 GB in bytes
		},
	}

	return result, nil
}
