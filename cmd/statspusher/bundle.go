package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/systemd"
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
	healthResults, errs := s.runHealthChecks(bundles)
	if len(errs) != 0 {
		return fmt.Errorf("bundle health check errors: %v", errs)
	}
	return s.sendBundleHeartbeats(healthResults, serial, ip)
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
		if err := acomm.Send(s.config.nodeDataURL(), req); err != nil {
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
	if err := acomm.Send(s.config.nodeDataURL(), req); err != nil {
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

func (s *statsPusher) sendBundleHeartbeats(bundles map[uint64]map[string]error, serial string, ip net.IP) error {
	errored := make([]uint64, 0, len(bundles))

	multiRequest := acomm.NewMultiRequest(s.tracker, s.config.requestTimeout())
	for bundle, healthErrors := range bundles {
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task:    "bundle-heartbeat",
			TaskURL: s.config.heartbeatURL(),
			Args: clusterconf.BundleHeartbeatArgs{
				ID:     bundle,
				Serial: serial,
				Node: clusterconf.BundleNode{
					IP:           ip,
					HealthErrors: healthErrors,
				},
			},
		})
		if err != nil {
			errored = append(errored, bundle)
			continue
		}
		if err := multiRequest.AddRequest(strconv.FormatUint(bundle, 10), req); err != nil {
			errored = append(errored, bundle)
			continue
		}
		if err := acomm.Send(s.config.nodeDataURL(), req); err != nil {
			multiRequest.RemoveRequest(req)
			errored = append(errored, bundle)
			continue
		}
	}

	responses := multiRequest.Responses()
	for name, resp := range responses {
		if resp.Error != nil {
			bundle, _ := strconv.ParseUint(name, 10, 64)
			errored = append(errored, bundle)
			break
		}
	}

	if len(errored) > 0 {
		return fmt.Errorf("one or more bundle heartbeats unsuccessful: %+v", errored)
	}
	return nil
}

func (s *statsPusher) runHealthChecks(bundles []*clusterconf.Bundle) (map[uint64]map[string]error, map[string]error) {
	multiRequest := acomm.NewMultiRequest(s.tracker, 0)

	requests := make(map[string]*acomm.Request)
	errors := make(map[string]error)
	for _, bundle := range bundles {
		for serviceID, service := range bundle.Services {
			for healthID, healthCheck := range service.HealthChecks {
				name := fmt.Sprintf("%d:%s:%s", bundle.ID, serviceID, healthID)
				req, err := acomm.NewRequest(acomm.RequestOptions{
					Task: healthCheck.Type,
					Args: healthCheck.Args,
				})
				if err != nil {
					errors[name] = fmt.Errorf("health check request creation for %s failed: %v", name, err)
					continue
				}
				requests[name] = req
			}
		}
	}

	for name, req := range requests {
		if err := multiRequest.AddRequest(name, req); err != nil {
			errors[name] = err
			continue
		}
		if err := acomm.Send(s.config.nodeDataURL(), req); err != nil {
			multiRequest.RemoveRequest(req)
			errors[name] = err
		}
	}

	responses := multiRequest.Responses()
	healthResults := make(map[uint64]map[string]error)
	for name, resp := range responses {
		nameParts := strings.Split(name, ":")
		bundleID, _ := strconv.ParseUint(nameParts[0], 10, 64)
		healthCheck := nameParts[1] + ":" + nameParts[2]
		if _, ok := healthResults[bundleID]; !ok {
			healthResults[bundleID] = make(map[string]error)
		}
		if resp.Error != nil {
			healthResults[bundleID][healthCheck] = fmt.Errorf("health check failed: %v", resp.Error)
		}
	}

	return healthResults, errors
}

func extractBundles(units []systemd.UnitStatus) []uint64 {
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
