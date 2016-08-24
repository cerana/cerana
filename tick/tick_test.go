package tick_test

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/test"
	"github.com/cerana/cerana/providers/metrics"
	"github.com/cerana/cerana/tick"
	"github.com/stretchr/testify/suite"
)

type Tick struct {
	suite.Suite
	config      *tick.Config
	configData  *tick.ConfigData
	configFile  *os.File
	tracker     *acomm.Tracker
	coordinator *test.Coordinator
	metrics     *metrics.MockMetrics
}

func TestTick(t *testing.T) {
	suite.Run(t, new(Tick))
}

func (s *Tick) SetupSuite() {
	var err error
	noError := s.Require().NoError

	logrus.SetLevel(logrus.FatalLevel)

	s.coordinator, err = test.NewCoordinator("")
	noError(err)

	nodeDataURL := s.coordinator.NewProviderViper().GetString("coordinator_url")
	s.configData = &tick.ConfigData{
		NodeDataURL:       nodeDataURL,
		ClusterDataURL:    nodeDataURL,
		HTTPResponseAddr:  ":54321",
		LogLevel:          "fatal",
		RequestTimeout:    "5s",
		TickInterval:      "1s",
		TickRetryInterval: "500ms",
	}

	s.config, _, _, s.configFile, err = newTestConfig(false, true, s.configData)
	noError(err, "failed to create config")
	noError(s.config.LoadConfig(), "failed to load config")

	tracker, err := acomm.NewTracker("", nil, nil, s.config.RequestTimeout())
	noError(err)
	s.tracker = tracker
	noError(s.tracker.Start())

	s.metrics = metrics.NewMockMetrics()
	s.coordinator.RegisterProvider(s.metrics)

	noError(s.coordinator.Start())
}

func (s *Tick) TearDownSuite() {
	s.coordinator.Stop()
	s.Require().NoError(s.coordinator.Cleanup())
	_ = os.Remove(s.configFile.Name())
	s.tracker.Stop()
}

func (s *Tick) TestRunTick() {
	output := make(chan time.Time, 10)

	tickFn := func(config tick.Configer, tracker *acomm.Tracker) error {
		output <- time.Now()
		return nil
	}

	stopChan, err := tick.RunTick(nil, tickFn)
	if !s.EqualError(err, "config required") {
		stopChan <- struct{}{}
		<-stopChan
	}

	stopChan, err = tick.RunTick(s.config, nil)
	if !s.EqualError(err, "tick function required") {
		stopChan <- struct{}{}
		<-stopChan
	}

	stopChan, err = tick.RunTick(s.config, tickFn)
	if !s.NoError(err) {
		return
	}

	time.Sleep(5 * time.Second)
	stopChan <- struct{}{}
	<-stopChan

	close(output)
	s.Len(output, 5)

	var prev time.Time
	for t := range output {
		if prev.IsZero() {
			prev = t
			continue
		}

		s.WithinDuration(t, prev, 1100*time.Millisecond)
		prev = t
	}

}

func (s *Tick) TestGetIP() {
	ip, err := tick.GetIP(s.config, s.tracker)
	if !s.NoError(err) {
		return
	}
	expected, _, _ := net.ParseCIDR(s.metrics.Data.Network.Interfaces[0].Addrs[0].Addr)
	s.EqualValues(expected, ip)
}
