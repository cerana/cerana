package tick

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/providers/metrics"
	"github.com/tylerb/graceful"
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

	httpServer, err := setupHTTPResponseServer(config.HTTPResponseURL(), tracker)
	if err != nil {
		tracker.Stop()
		return nil, err
	}

	go func() {
		defer func() { stopChan <- struct{}{} }()
		defer tracker.Stop()
		defer stopHTTP(httpServer, config.RequestTimeout())

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

			logrus.WithField("interval", interval).Debug("waiting to run tick")

			select {
			case <-stopChan:
				logrus.Debug("stopping tick")
				return
			case sig := <-sigChan:
				logrus.WithField("signal", sig).Debug("signal received, stopping tick")
				return
			case lastStart = <-time.After(interval):
				logrus.Debug("running tick")
				if err = tick(config, tracker); err != nil {
					logrus.WithField("error", err).Error("tick error")
				} else {
					logrus.Debug("successful tick")
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

func setupHTTPResponseServer(responseURL *url.URL, tracker *acomm.Tracker) (*graceful.Server, error) {
	if responseURL == nil {
		return nil, nil
	}

	// setup external response handler
	mux := http.NewServeMux()
	mux.HandleFunc(responseURL.Path, tracker.ProxyExternalHandler)
	httpServer := &graceful.Server{
		Server: &http.Server{
			Addr:    responseURL.Host,
			Handler: mux,
		},
		NoSignalHandling: true,
	}

	// run the http server and wait briefly to make sure it started without error
	errChan := make(chan error, 1)
	go func() {
		errChan <- errors.Wrapv(httpServer.ListenAndServe(), map[string]interface{}{"addr": responseURL.Host, "path": responseURL.Path})
	}()
	select {
	case err := <-errChan:
		return nil, err
	case <-time.After(time.Second):
		// started cleanly
	}

	return httpServer, nil
}

func stopHTTP(httpServer *graceful.Server, timeout time.Duration) {
	if httpServer == nil {
		return
	}

	httpStop := httpServer.StopChan()
	httpServer.Stop(timeout)
	<-httpStop
}
