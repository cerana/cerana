package clusterconf_test

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/test"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/kv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type clusterConf struct {
	suite.Suite
	coordinator  *test.Coordinator
	config       *clusterconf.Config
	clusterConf  *clusterconf.ClusterConf
	tracker      *acomm.Tracker
	viper        *viper.Viper
	responseHook *url.URL
	kv           *kv.Mock
}

func TestClusterConf(t *testing.T) {
	suite.Run(t, new(clusterConf))
}

func (s *clusterConf) SetupSuite() {
	var err error
	s.coordinator, err = test.NewCoordinator("")
	s.Require().NoError(err)

	s.responseHook, _ = url.ParseRequestURI("unix:///tmp/foobar")

	v := s.coordinator.NewProviderViper()
	v.Set("dataset_ttl", time.Minute)
	v.Set("bundle_ttl", time.Minute)
	v.Set("node_ttl", time.Minute)
	flagset := pflag.NewFlagSet("clusterconf", pflag.PanicOnError)
	config := clusterconf.NewConfig(flagset, v)
	s.Require().NoError(flagset.Parse([]string{}))
	s.Require().NoError(config.LoadConfig())
	s.Require().NoError(config.SetupLogging())
	s.config = config
	s.viper = v

	s.tracker, err = acomm.NewTracker(filepath.Join(s.coordinator.SocketDir, "tracker.sock"), nil, nil, 5*time.Second)
	s.Require().NoError(err)
	s.Require().NoError(s.tracker.Start())

	s.clusterConf = clusterconf.New(config, s.tracker)

	v = s.coordinator.NewProviderViper()
	flagset = pflag.NewFlagSet("kv", pflag.PanicOnError)
	kvConfig := provider.NewConfig(flagset, v)
	s.Require().NoError(flagset.Parse([]string{}))
	s.Require().NoError(kvConfig.LoadConfig())
	s.kv, err = kv.NewMock(kvConfig, s.coordinator.ProviderTracker())
	s.Require().NoError(err)
	s.coordinator.RegisterProvider(s.kv)

	s.Require().NoError(s.coordinator.Start())
}

func (s *clusterConf) TearDownTest() {
	s.Require().NoError(s.clearData())
}

func (s *clusterConf) TearDownSuite() {
	s.coordinator.Stop()
	s.kv.Stop()
	s.tracker.Stop()
	s.Require().NoError(s.coordinator.Cleanup())
}

func (s *clusterConf) TestRegisterTasks() {
	server, err := provider.NewServer(s.config.Config)
	s.Require().NoError(err)

	s.clusterConf.RegisterTasks(server)

	s.True(len(server.RegisteredTasks()) > 0)
}

func (s *clusterConf) loadData(data map[string]interface{}) (map[string]uint64, error) {
	multiRequest := acomm.NewMultiRequest(s.tracker, 0)
	requests := make(map[string]*acomm.Request)
	for key, v := range data {
		value, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "kv-update",
			Args: kv.UpdateArgs{
				Key:   key,
				Value: string(value),
			},
		})
		if err != nil {
			return nil, err
		}
		requests[key] = req
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

	responses := multiRequest.Responses()
	indexes := make(map[string]uint64)
	for name := range requests {
		resp, ok := responses[name]
		if !ok {
			return nil, fmt.Errorf("failed to send request: %s", name)
		}
		if resp.Error != nil {
			return nil, fmt.Errorf("request failed: %s: %s", name, resp.Error)
		}
		var result kv.UpdateReturn
		if err := resp.UnmarshalResult(&result); err != nil {
			return nil, err
		}
		indexes[name] = result.Index
	}

	return indexes, nil
}

func (s *clusterConf) clearData() error {
	respChan := make(chan *acomm.Response, 1)
	defer close(respChan)
	rh := func(_ *acomm.Request, resp *acomm.Response) {
		respChan <- resp
	}

	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:         "kv-delete",
		ResponseHook: s.tracker.URL(),
		Args: kv.DeleteArgs{
			Key:       "",
			Recursive: true,
		},
		SuccessHandler: rh,
		ErrorHandler:   rh,
	})
	if err != nil {
		return err
	}
	if err := s.tracker.TrackRequest(req, 0); err != nil {
		return err
	}
	if err := acomm.Send(s.config.CoordinatorURL(), req); err != nil {
		s.tracker.RemoveRequest(req)
		return err
	}

	resp := <-respChan
	return resp.Error
}
