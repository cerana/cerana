package provider_test

import (
	"io/ioutil"
	"net/url"
	"os"
	"syscall"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/provider"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/suite"
)

func TestServer(t *testing.T) {
	suite.Run(t, new(ServerSuite))
}

type ServerSuite struct {
	suite.Suite
	config     *provider.Config
	configData *provider.ConfigData
	server     *provider.Server
}

func (s *ServerSuite) SetupSuite() {
	log.SetLevel(log.FatalLevel)

	socketDir, err := ioutil.TempDir("", "providerTest-")
	s.Require().NoError(err, "failed to create socket dir")

	s.configData = &provider.ConfigData{
		SocketDir:       socketDir,
		ServiceName:     uuid.New(),
		CoordinatorURL:  "http://localhost:8080/",
		DefaultPriority: 43,
		LogLevel:        "fatal",
		DefaultTimeout:  100,
		RequestTimeout:  10,
		Tasks: map[string]*provider.TaskConfigData{
			"foobar": {
				Priority: 56,
				Timeout:  64,
			},
		},
	}

	s.config, _, _, _, err = newConfig(true, false, s.configData)
	s.Require().NoError(err, "failed to create config")
	s.Require().NoError(s.config.LoadConfig(), "failed to load config")
}

func (s *ServerSuite) SetupTest() {
	var err error
	s.server, err = provider.NewServer(s.config)
	s.Require().NoError(err, "failed to create server")
	s.Require().NotNil(s.server, "failed to create server")
}

func (s *ServerSuite) TearDownSuite() {
	_ = os.RemoveAll(s.configData.SocketDir)
}

func (s *ServerSuite) TestNewServer() {
	configInvalid := provider.NewConfig(nil, nil)

	server, err := provider.NewServer(configInvalid)
	s.Nil(server, "should not create server with invalid config")
	s.Error(err, "should error with invalid config")
}

func (s *ServerSuite) TestTracker() {
	s.NotNil(s.server.Tracker())
}

func (s *ServerSuite) TestRegisterTask() {
	handler := func(a *acomm.Request) (interface{}, *url.URL, error) {
		return nil, nil, nil
	}
	s.server.RegisterTask("foobar", handler)

	tasks := s.server.RegisteredTasks()
	if !s.Len(tasks, 1, "should be one registered task") {
		return
	}
	s.Equal("foobar", tasks[0], "should be registered under correct name")
}

func (s *ServerSuite) TestStartHandleStop() {
	// Start
	taskHandler := func(a *acomm.Request) (interface{}, *url.URL, error) {
		return nil, nil, nil
	}
	s.server.RegisterTask("foobar", taskHandler)

	if !s.NoError(s.server.Start(), "failed to start server") {
		return
	}
	time.Sleep(time.Second)

	// Stop
	defer s.server.Stop()

	// Handle request
	tracker := s.server.Tracker()
	handled := make(chan struct{})
	respHandler := func(req *acomm.Request, resp *acomm.Response) {
		close(handled)
	}
	req, err := acomm.NewRequest(&acomm.RequestOptions{
		Task:           "foobar",
		ResponseHook:   tracker.URL(),
		SuccessHandler: respHandler,
		ErrorHandler:   respHandler,
	})
	s.Require().NoError(err)

	providerSocket, _ := url.ParseRequestURI("unix://" + s.server.TaskSocketPath("foobar"))
	if !s.NoError(s.server.Tracker().TrackRequest(req, 5*time.Second)) {
		return
	}
	if !s.NoError(acomm.Send(providerSocket, req)) {
		return
	}
	<-handled
}

func (s *ServerSuite) TestStopOnSignal() {
	selfProcess, err := os.FindProcess(os.Getpid())
	if !s.NoError(err, "couldn't find this process") {
		return
	}

	if !s.NoError(s.server.Start(), "failed to start server") {
		return
	}

	stopSignal := syscall.SIGUSR1

	done := make(chan struct{})
	go func() {
		s.server.StopOnSignal(stopSignal)
		close(done)
	}()

	time.Sleep(time.Second)
	_ = selfProcess.Signal(stopSignal)

	<-done
}
