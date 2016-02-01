package simple

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/acomm"
	"github.com/mistifyio/provider"
)

// Simple is a simple provider implementation.
type Simple struct {
	config  *provider.Config
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

// DelayedRespArgs are arguments for the DelayedResp handler.
type DelayedRespArgs struct {
	Delay time.Duration `json:"delay"`
}

// DelayedRespResult is the result data for the DelayedResp handler.
type DelayedRespResult struct {
	Delay       time.Duration `json:"delay"`
	ReceivedAt  time.Time     `json:"received_at"`
	RespondedAt time.Time     `json:"responded_at"`
}

// NewSimple creates a new instance of Simple.
func NewSimple(config *provider.Config, tracker *acomm.Tracker) *Simple {
	return &Simple{
		config:  config,
		tracker: tracker,
	}
}

// RegisterTasks registers all of Simple's task handlers with the server.
func (s *Simple) RegisterTasks(server *provider.Server) {
	server.RegisterTask("SystemStatus", s.SystemStatus)
	server.RegisterTask("CPUInfo", s.CPUInfo)
	server.RegisterTask("DiskInfo", s.DiskInfo)
	server.RegisterTask("StreamEcho", s.StreamEcho)
	server.RegisterTask("DelayedResp", s.DelayedResp)
}

// SystemStatus is a task handler to retrieve info look up and return system
// information. It depends on and makes requests for several other tasks.
func (s *Simple) SystemStatus(req *acomm.Request) (interface{}, *url.URL, error) {
	var args SystemStatusArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.GuestID == "" {
		return nil, nil, errors.New("missing guest_id")
	}

	// Prepare multiple requests
	multiRequest := acomm.NewMultiRequest(s.tracker, 0)

	cpuReq, err := acomm.NewRequest("CPUInfo", s.tracker.URL().String(), &CPUInfoArgs{GuestID: args.GuestID}, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	diskReq, err := acomm.NewRequest("DiskInfo", s.tracker.URL().String(), &DiskInfoArgs{GuestID: args.GuestID}, nil, nil)
	if err != nil {
		return nil, nil, err
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

	return result, nil, nil
}

// CPUInfo is a task handler to retrieve information about CPUs.
func (s *Simple) CPUInfo(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CPUInfoArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.GuestID == "" {
		return nil, nil, errors.New("missing guest_id")
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
	return result, nil, nil
}

// DiskInfo is a task handler to retrieve information about disks.
func (s *Simple) DiskInfo(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CPUInfoArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.GuestID == "" {
		return nil, nil, errors.New("missing guest_id")
	}

	result := &DiskInfoResult{
		&DiskInfo{
			Device: "vda1",
			Size:   10 * (1024 * 1024 * 1024), // 10 GB in bytes
		},
	}

	return result, nil, nil
}

// StreamEcho is a task handler to echo input back via streaming data.
func (s *Simple) StreamEcho(req *acomm.Request) (interface{}, *url.URL, error) {
	src := ioutil.NopCloser(bytes.NewReader(*req.Args))
	socketDir := filepath.Join(
		s.config.SocketDir(),
		"streams",
		"StreamEcho",
		s.config.ServiceName())
	addr, err := s.tracker.NewStreamUnix(socketDir, src)

	return nil, addr, err
}

// DelayedResp is a task handler that waits a specified time before returning.
func (s *Simple) DelayedResp(req *acomm.Request) (interface{}, *url.URL, error) {
	var args DelayedRespArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Delay <= 0 {
		return nil, nil, errors.New("delay must be positive")
	}

	start := time.Now()
	time.Sleep(time.Second * args.Delay)

	result := &DelayedRespResult{
		Delay:       args.Delay,
		ReceivedAt:  start,
		RespondedAt: time.Now(),
	}

	return result, nil, nil
}
