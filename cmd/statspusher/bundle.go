package main

import (
	"errors"
	"net"
	"strconv"
	"strings"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/systemd"
	"github.com/coreos/go-systemd/dbus"
	"github.com/pborman/uuid"
	"github.com/shirou/gopsutil/host"
)

func (s *statsPusher) bundleHeartbeats() error {
	serial, err := s.getSerial()
	if err != nil {
		return err
	}
	ip, err := s.getIP()
	if err != nil {
		return err
	}
	bundles, err := s.getBundles()
	if err != nil {
		return err
	}
	healthy, err := s.runHealthChecks(bundles)
	if err != nil {
		//log
	}
	return s.sendBundleHeartbeats(healthy, serial, ip)
}

func (s *statsPusher) getBundles() ([]*clusterconf.Bundle, error) {
	requests := make(map[string]*acomm.Request)
	localReq, err := acomm.NewRequest(acomm.RequestOptions{Task: "systemd-list"})
	if err != nil {
		return nil, err
	}
	requests["local"] = localReq
	knownReq, err := acomm.NewRequest(acomm.RequestOptions{
		Task:    "list-bundles",
		TaskURL: s.config.heartbeatURL(),
	})
	requests["known"] = knownReq

	multiRequest := acomm.NewMultiRequest(s.tracker, s.config.requestTimeout())
	for name, req := range requests {
		if err := multiRequest.AddRequest(name, req); err != nil {
			break
		}
		if err := acomm.Send(s.config.coordinatorURL(), req); err != nil {
			multiRequest.RemoveRequest(req)
			break
		}
	}

	responses := multiRequest.Responses()

	var localUnits systemd.ListResult
	if err := responses["local"].UnmarshalResult(&localUnits); err != nil {
		return nil, err
	}
	localBundles := extractBundles(localUnits.Units)
	var knownBundles clusterconf.BundleListResult
	if err := responses["known"].UnmarshalResult(&knownBundles); err != nil {
		return nil, err
	}

	bundles := make([]*clusterconf.Bundle, 0, len(localBundles))
	for _, local := range localBundles {
		for _, known := range knownBundles.Bundles {
			if known.ID == local {
				bundles = append(bundles, known)
				break
			}
		}
	}

	return bundles, nil
}

func (s *statsPusher) getSerial() (string, error) {
	doneChan := make(chan *acomm.Response, 1)
	rh := func(_ *acomm.Request, resp *acomm.Response) {
		doneChan <- resp
	}
	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:           "metrics-host",
		ResponseHook:   s.tracker.URL(),
		SuccessHandler: rh,
		ErrorHandler:   rh,
	})
	if err != nil {
		return "", err
	}
	if err := s.tracker.TrackRequest(req, s.config.requestTimeout()); err != nil {
		return "", err
	}
	if err := acomm.Send(s.config.coordinatorURL(), req); err != nil {
		return "", err
	}

	resp := <-doneChan
	if resp.Error != nil {
		return "", resp.Error
	}

	var data host.InfoStat
	if err := resp.UnmarshalResult(&data); err != nil {
		return "", err
	}

	return data.Hostname, nil
}

func (s *statsPusher) sendBundleHeartbeats(bundles []uint64, serial string, ip net.IP) error {
	var errored bool

	multiRequest := acomm.NewMultiRequest(s.tracker, s.config.requestTimeout())
	for _, bundle := range bundles {
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task:    "bundle-heartbeat",
			TaskURL: s.config.heartbeatURL(),
			Args: clusterconf.BundleHeartbeatArgs{
				ID:     bundle,
				Serial: serial,
				IP:     ip,
			},
		})
		if err != nil {
			errored = true
			continue
		}
		if err := multiRequest.AddRequest(strconv.FormatUint(bundle, 10), req); err != nil {
			errored = true
			continue
		}
		if err := acomm.Send(s.config.coordinatorURL(), req); err != nil {
			multiRequest.RemoveRequest(req)
			errored = true
			continue
		}
	}
	responses := multiRequest.Responses()
	for _, resp := range responses {
		if resp.Error != nil {
			errored = true
			break
		}
	}

	if errored {
		return errors.New("one or more bundle heartbeats unsuccessful")
	}
	return nil
}

// TODO: Make this actually run health checks
func (s *statsPusher) runHealthChecks(bundles []*clusterconf.Bundle) ([]uint64, error) {
	healthy := make([]uint64, len(bundles))
	for i, bundle := range bundles {
		healthy[i] = bundle.ID
	}
	return healthy, nil
}

func extractBundles(units []dbus.UnitStatus) []uint64 {
	dedupe := make(map[uint64]bool)
	for _, unit := range units {
		// bundleID:serviceID
		parts := strings.Split(unit.Name, ":")
		bundleID, err := strconv.ParseUint(parts[0], 10, 64)
		if err == nil && len(parts) == 2 && uuid.Parse(parts[1]) != nil {
			dedupe[bundleID] = true
		}
	}
	ids := make([]uint64, 0, len(dedupe))
	for id := range dedupe {
		ids = append(ids, id)
	}
	return ids
}
