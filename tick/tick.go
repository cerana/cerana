package tick

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/providers/metrics"
)

// ActionFn is a function that can be run on a tick interval.
type ActionFn func(Configer, *acomm.Tracker) error

// RunTick runs the supplied function on a configured interval. Putting an
// entry into the returned chan, as well as sending an os signal, can be used
// to stop the tick.
func RunTick(config Configer, tick ActionFn) (chan struct{}, error) {
	if config == nil {
		return nil, errors.New("config required")
	}
	if tick == nil {
		return nil, errors.New("tick function required")
	}

	// allow tick to be stopped cleanly with signal or manually
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	stopChan := make(chan struct{}, 1)

	// setup response tracker
	tracker, err := acomm.NewTracker("", nil, nil, config.RequestTimeout())
	if err != nil {
		return nil, err
	}
	if err = tracker.Start(); err != nil {
		return nil, err
	}

	go func() {
		defer func() { stopChan <- struct{}{} }()
		defer tracker.Stop()

		lastStart := time.Time{}
		for {
			// Try to start fn at the desired time interval, accounting for
			// the time it takes for fn to complete while also preventing fn
			// from running more than once at a time.
			interval := config.TickInterval()
			if err != nil && config.TickRetryInterval() > 0 {
				interval = config.TickRetryInterval()
			}
			since := time.Since(lastStart)
			if since < interval {
				interval = interval - since
			} else {
				interval = time.Duration(0)
			}

			select {
			case <-stopChan:
				return
			case <-sigChan:
				return
			case lastStart = <-time.After(interval):
				if err = tick(config, tracker); err != nil {
					logrus.WithField("error", err).Error("tick error")
				}
			}
		}
	}()
	return stopChan, nil
}

// GetIP retrieves the ip of the current node, very often needed for ticks.
func GetIP(config Configer, tracker *acomm.Tracker) (net.IP, error) {
	opts := acomm.RequestOptions{
		Task: "metrics-network",
	}
	resp, err := tracker.SyncRequest(config.NodeDataURL(), opts, config.RequestTimeout())
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, errors.ResetStack(resp.Error)
	}

	var data metrics.NetworkResult
	if err := resp.UnmarshalResult(&data); err != nil {
		return nil, err
	}
	for _, iface := range data.Interfaces {
		for _, ifaceAddr := range iface.Addrs {
			ip, _, _ := net.ParseCIDR(ifaceAddr.Addr)
			if ip != nil && !ip.IsLoopback() {
				return ip, nil
			}
		}
	}
	return nil, errors.New("no suitable IP found")
}
