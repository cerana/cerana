package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
)

type statsPusher struct {
	config   *config
	tracker  *acomm.Tracker
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func newStatsPusher(c *config) (*statsPusher, error) {
	tracker, err := acomm.NewTracker("", nil, nil, c.requestTimeout())
	if err != nil {
		return nil, err
	}

	return &statsPusher{
		config:   c,
		tracker:  tracker,
		stopChan: make(chan struct{}),
	}, nil
}

func (s *statsPusher) run() error {
	if err := s.tracker.Start(); err != nil {
		return err
	}
	s.startHeartbeat("node", s.nodeHeartbeat, s.config.nodeTTL())
	s.startHeartbeat("dataset", s.datasetHeartbeats, s.config.datasetTTL())
	s.startHeartbeat("bundle", s.bundleHeartbeats, s.config.bundleTTL())
	return nil
}

func (s *statsPusher) startHeartbeat(name string, fn func() error, desiredInterval time.Duration) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		lastStart := time.Time{}
		for {
			// Try to start fn at the desired time interval, accounting for
			// time time it takes for fn to complete while also preventing fn
			// from running more than once at a time.
			interval := desiredInterval
			since := time.Since(lastStart)
			if since < interval {
				interval = since
			}
			timeChan := time.After(interval)

			select {
			case <-s.stopChan:
				break
			case lastStart = <-timeChan:
				if err := fn(); err != nil {
					logrus.WithFields(logrus.Fields{
						"name":  name,
						"error": err,
					}).Error("error while running heartbeat")
				}
			}
		}
	}()
}

func (s *statsPusher) stop() {
	close(s.stopChan)
	s.wg.Wait()

	s.tracker.Stop()
}

func (s *statsPusher) stopOnSignal(signals ...os.Signal) {
	if len(signals) == 0 {
		signals = []os.Signal{os.Interrupt, os.Kill, syscall.SIGTERM}
	}

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, signals...)
	sig := <-sigChan
	logrus.WithFields(logrus.Fields{
		"signal": sig,
	}).Info("signal received, stopping")

	s.stop()
}
