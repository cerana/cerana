package main

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/service"
	"github.com/cerana/cerana/tick"
	"github.com/shirou/gopsutil/host"
)

func bundleHeartbeats(config tick.Configer, tracker *acomm.Tracker) error {
	serial, err := getSerial(config, tracker)
	if err != nil {
		return err
	}
	ip, err := tick.GetIP(config, tracker)
	if err != nil {
		return err
	}
	bundles, err := getBundles(config, tracker)
	if err != nil {
		return err
	}
	healthResults, errs := runHealthChecks(config, tracker, bundles)
	if len(errs) != 0 {
		return errors.Newv("bundle health check errors", map[string]interface{}{"errors": errs})
	}
	return sendBundleHeartbeats(config, tracker, healthResults, serial, ip)
}

func getBundles(config tick.Configer, tracker *acomm.Tracker) ([]*clusterconf.Bundle, error) {
	requests := map[string]struct {
		task     string
		url      *url.URL
		respData interface{}
	}{
		"local": {task: "service-list", url: config.NodeDataURL(), respData: &service.ListResult{}},
		"known": {task: "list-bundles", url: config.ClusterDataURL(), respData: &clusterconf.BundleListResult{}},
	}

	multiRequest := acomm.NewMultiRequest(tracker, config.RequestTimeout())
	for name, args := range requests {
		req, err := acomm.NewRequest(acomm.RequestOptions{Task: args.task})
		if err != nil {
			return nil, err
		}
		if err := multiRequest.AddRequest(name, req); err != nil {
			return nil, err
		}
		if err := acomm.Send(args.url, req); err != nil {
			multiRequest.RemoveRequest(req)
			return nil, err
		}

	}

	responses := multiRequest.Responses()
	for name, args := range requests {
		resp := responses[name]
		if resp.Error != nil {
			return nil, errors.ResetStack(resp.Error)
		}
		if err := resp.UnmarshalResult(args.respData); err != nil {
			return nil, err
		}
	}
	localBundles := extractBundles(requests["local"].respData.(*service.ListResult).Services)
	knownBundles := requests["known"].respData.(*clusterconf.BundleListResult).Bundles

	bundles := make([]*clusterconf.Bundle, 0, len(localBundles))
	for _, local := range localBundles {
		// Attempt to add the known bundle with service and healthcheck info.
		found := false
		for _, known := range knownBundles {
			if known.ID == local {
				bundles = append(bundles, known)
				found = true
				break
			}
		}

		// If not found, add an entry anyway so it is tracked by heartbeat.
		// Something will later clean the untracked bundle up.
		if !found {
			bundles = append(bundles, &clusterconf.Bundle{ID: local})
		}
	}

	return bundles, nil
}

func getSerial(config tick.Configer, tracker *acomm.Tracker) (string, error) {
	doneChan := make(chan *acomm.Response, 1)
	rh := func(_ *acomm.Request, resp *acomm.Response) {
		doneChan <- resp
	}
	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:           "metrics-host",
		ResponseHook:   tracker.URL(),
		SuccessHandler: rh,
		ErrorHandler:   rh,
	})
	if err != nil {
		return "", err
	}
	if err := tracker.TrackRequest(req, config.RequestTimeout()); err != nil {
		return "", err
	}
	if err := acomm.Send(config.NodeDataURL(), req); err != nil {
		return "", err
	}

	resp := <-doneChan
	if resp.Error != nil {
		return "", errors.ResetStack(resp.Error)
	}

	var data host.InfoStat
	if err := resp.UnmarshalResult(&data); err != nil {
		return "", err
	}

	return data.Hostname, nil
}

func sendBundleHeartbeats(config tick.Configer, tracker *acomm.Tracker, bundles map[uint64]map[string]error, serial string, ip net.IP) error {
	errored := make([]uint64, 0, len(bundles))

	multiRequest := acomm.NewMultiRequest(tracker, config.RequestTimeout())
	for bundle, healthErrors := range bundles {
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "bundle-heartbeat",
			Args: clusterconf.BundleHeartbeatArgs{
				ID:           bundle,
				Serial:       serial,
				IP:           ip,
				HealthErrors: healthErrors,
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
		if err := acomm.Send(config.ClusterDataURL(), req); err != nil {
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
		return errors.Newv("one or more bundle heartbeats unsuccessful", map[string]interface{}{"errors": errored})
	}
	return nil
}

func runHealthChecks(config tick.Configer, tracker *acomm.Tracker, bundles []*clusterconf.Bundle) (map[uint64]map[string]error, map[string]error) {
	multiRequest := acomm.NewMultiRequest(tracker, 0)

	requests := make(map[string]*acomm.Request)
	errs := make(map[string]error)
	for _, bundle := range bundles {
		for serviceID, service := range bundle.Services {
			for healthID, healthCheck := range service.HealthChecks {
				name := fmt.Sprintf("%d:%s:%s", bundle.ID, serviceID, healthID)
				req, err := acomm.NewRequest(acomm.RequestOptions{
					Task: healthCheck.Type,
					Args: healthCheck.Args,
				})
				if err != nil {
					errs[name] = errors.Wrapv(err, map[string]interface{}{"name": name}, "health check request creation failed")
					continue
				}
				requests[name] = req
			}
		}
	}

	for name, req := range requests {
		if err := multiRequest.AddRequest(name, req); err != nil {
			errs[name] = err
			continue
		}
		if err := acomm.Send(config.NodeDataURL(), req); err != nil {
			multiRequest.RemoveRequest(req)
			errs[name] = err
		}
	}

	responses := multiRequest.Responses()
	healthResults := make(map[uint64]map[string]error)
	for _, bundle := range bundles {
		healthResults[bundle.ID] = make(map[string]error)
	}

	for name, resp := range responses {
		nameParts := strings.Split(name, ":")
		bundleID, _ := strconv.ParseUint(nameParts[0], 10, 64)
		healthCheck := nameParts[1] + ":" + nameParts[2]
		if resp.Error != nil {
			healthResults[bundleID][healthCheck] = errors.ResetStack(resp.Error)
		}
	}

	return healthResults, errs
}

func extractBundles(services []service.Service) []uint64 {
	dedupe := make(map[uint64]bool)
	for _, service := range services {
		dedupe[service.BundleID] = true
	}
	ids := make([]uint64, 0, len(dedupe))
	for id := range dedupe {
		ids = append(ids, id)
	}
	return ids
}
