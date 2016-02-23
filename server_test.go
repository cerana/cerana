package coordinator_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/acomm"
	"github.com/mistifyio/coordinator"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/suite"
)

func TestServer(t *testing.T) {
	suite.Run(t, new(ServerSuite))
}

type ServerSuite struct {
	suite.Suite
	config     *coordinator.Config
	configData *coordinator.ConfigData
	server     *coordinator.Server
}

func (s *ServerSuite) SetupSuite() {
	log.SetLevel(log.FatalLevel)

	socketDir, err := ioutil.TempDir("", "coordinatorTest-")
	s.Require().NoError(err, "failed to create socket dir")

	s.configData = &coordinator.ConfigData{
		SocketDir:      socketDir,
		ServiceName:    uuid.New(),
		ExternalPort:   45678,
		RequestTimeout: 5,
		LogLevel:       "fatal",
	}

	s.config, _, _, _, err = newConfig(true, false, s.configData)
	s.Require().NoError(err, "failed to create config")
	s.Require().NoError(s.config.LoadConfig(), "failed to load config")
}

func (s *ServerSuite) SetupTest() {
	var err error
	s.server, err = coordinator.NewServer(s.config)
	s.Require().NoError(err, "failed to create server")
	s.Require().NotNil(s.server, "failed to create server")
}

func (s *ServerSuite) TearDownSuite() {
	_ = os.RemoveAll(s.configData.SocketDir)
}

func (s *ServerSuite) TestNewServer() {
	configInvalid := coordinator.NewConfig(nil, nil)

	server, err := coordinator.NewServer(configInvalid)
	s.Nil(server, "should not create server with invalid config")
	s.Error(err, "should error with invalid config")
}

func (s *ServerSuite) TestStartHandleHTTPStop() {
	// Start
	if !s.NoError(s.server.Start(), "failed to start server") {
		return
	}
	time.Sleep(time.Second)

	// Stop
	defer s.server.Stop()

	// Handle request
	failed := make(chan struct{})
	handled := make(chan struct{})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(handled)
	}))
	defer ts.Close()

	taskListener := s.createTaskListener(failed)
	if taskListener == nil {
		return
	}
	defer taskListener.Stop(0)

	req, _ := acomm.NewRequest("foobar", ts.URL, struct{}{}, nil, nil)
	coordinatorURL, _ := url.ParseRequestURI(fmt.Sprintf("http://localhost:%v", s.configData.ExternalPort))
	if !s.NoError(acomm.Send(coordinatorURL, req)) {
		return
	}

	success := false
	select {
	case <-handled:
		success = true
	case <-failed:
	}

	s.True(success)
}

func (s *ServerSuite) TestStartHandleUnixStop() {
	// Start
	if !s.NoError(s.server.Start(), "failed to start server") {
		return
	}
	time.Sleep(time.Second)

	// Stop
	defer s.server.Stop()

	// Handle request
	failed := make(chan struct{})
	handled := make(chan struct{})

	responseListener := acomm.NewUnixListener(filepath.Join(s.configData.SocketDir, "testResponse.sock"), 0)
	if !s.NoError(responseListener.Start(), "failed to start task listener") {
		return
	}
	defer responseListener.Stop(0)
	go func() {
		conn := responseListener.NextConn()
		if conn == nil {
			return
		}
		defer responseListener.DoneConn(conn)

		close(handled)
		time.Sleep(100 * time.Millisecond)
	}()

	taskListener := s.createTaskListener(failed)
	if taskListener == nil {
		return
	}
	defer taskListener.Stop(0)

	req, _ := acomm.NewRequest("foobar", responseListener.URL().String(), struct{}{}, nil, nil)
	coordinatorURL, _ := url.ParseRequestURI("unix://" + filepath.Join(s.config.SocketDir(), "coordinator", s.config.ServiceName()+".sock"))
	if !s.NoError(acomm.Send(coordinatorURL, req)) {
		return
	}

	success := false
	select {
	case <-handled:
		success = true
	case <-failed:
	}

	s.True(success)
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

func (s *ServerSuite) createTaskListener(failed chan struct{}) *acomm.UnixListener {
	taskListener := acomm.NewUnixListener(filepath.Join(s.configData.SocketDir, "foobar", "test.sock"), 0)
	if !s.NoError(taskListener.Start(), "failed to start task listener") {
		return nil
	}

	go func() {
		conn := taskListener.NextConn()
		if conn == nil {
			return
		}
		defer taskListener.DoneConn(conn)

		req := &acomm.Request{}
		if err := acomm.UnmarshalConnData(conn, req); err != nil {
			close(failed)
			return
		}

		// Respond to the initial request
		resp, _ := acomm.NewResponse(req, nil, nil, nil)
		if err := acomm.SendConnData(conn, resp); err != nil {
			close(failed)
			return
		}

		// Response to hook
		resp, _ = acomm.NewResponse(req, nil, nil, nil)
		if err := req.Respond(resp); err != nil {
			close(failed)
			return
		}
	}()

	time.Sleep(time.Second)
	return taskListener
}
